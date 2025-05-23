package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"sync"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/shopspring/decimal"
	"service/util"
)

type TxnFunc func(txn *util.TiDBSqlTx) error

const (
	ErrWriteConflict      = 9007 // Transactions in TiKV encounter write conflicts.
	ErrInfoSchemaChanged  = 8028 // table schema changes
	ErrForUpdateCantRetry = 8002 // "SELECT FOR UPDATE" commit conflict
	ErrTxnRetryable       = 8022 // The transaction commit fails and has been rolled back
)

const retryTimes = 5

var retryErrorCodeSet = map[uint16]interface{}{
	ErrWriteConflict:      nil,
	ErrInfoSchemaChanged:  nil,
	ErrForUpdateCantRetry: nil,
	ErrTxnRetryable:       nil,
}

func runTxn(db *sql.DB, optimistic bool, optimisticRetryTimes int, txnFunc TxnFunc) {
	txn, err := util.TiDBSqlBegin(db, !optimistic)
	if err != nil {
		panic(err)
	}

	err = txnFunc(txn)
	if err != nil {
		txn.Rollback()
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && optimistic && optimisticRetryTimes != 0 {
			if _, retryableError := retryErrorCodeSet[mysqlErr.Number]; retryableError {
				fmt.Printf("[runTxn] got a retryable error, rest time: %d\n", optimisticRetryTimes-1)
				runTxn(db, optimistic, optimisticRetryTimes-1, txnFunc)
				return
			}
		}

		fmt.Printf("[runTxn] got an error, rollback: %+v\n", err)
	} else {
		err = txn.Commit()
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && optimistic && optimisticRetryTimes != 0 {
			if _, retryableError := retryErrorCodeSet[mysqlErr.Number]; retryableError {
				fmt.Printf("[runTxn] got a retryable error, rest time: %d\n", optimisticRetryTimes-1)
				runTxn(db, optimistic, optimisticRetryTimes-1, txnFunc)
				return
			}
		}

		if err == nil {
			fmt.Println("[runTxn] commit success")
		}
	}
}

func prepareData(db *sql.DB, optimistic bool) {
	runTxn(db, optimistic, retryTimes, func(txn *util.TiDBSqlTx) error {
		publishedAt, err := time.Parse("2006-01-02 15:04:05", "2018-09-01 00:00:00")
		if err != nil {
			return err
		}

		if err = createBook(txn, 1, "Designing Data-Intensive Application",
			"Science & Technology", publishedAt, decimal.NewFromInt(100), 10); err != nil {
			return err
		}

		if err = createUser(txn, 1, "Bob", decimal.NewFromInt(10000)); err != nil {
			return err
		}

		if err = createUser(txn, 2, "Alice", decimal.NewFromInt(10000)); err != nil {
			return err
		}

		return nil
	})
}

func buyPessimistic(db *sql.DB, goroutineID, orderID, bookID, userID, amount int) {
	txnComment := fmt.Sprintf("/* txn %d */ ", goroutineID)
	if goroutineID != 1 {
		txnComment = "\t" + txnComment
	}

	fmt.Printf("\nuser %d try to buy %d books(id: %d)\n", userID, amount, bookID)

	runTxn(db, false, retryTimes, func(txn *util.TiDBSqlTx) error {
		time.Sleep(time.Second)

		// read the price of book
		selectBookForUpdate := "select `price` from books where id = ? for update"
		bookRows, err := txn.Query(selectBookForUpdate, bookID)
		if err != nil {
			return err
		}
		fmt.Println(txnComment + selectBookForUpdate + " successful")
		defer bookRows.Close()

		price := decimal.NewFromInt(0)
		if bookRows.Next() {
			err = bookRows.Scan(&price)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("book ID not exist")
		}
		bookRows.Close()

		// update book
		updateStock := "update `books` set stock = stock - ? where id = ? and stock - ? >= 0"
		result, err := txn.Exec(updateStock, amount, bookID, amount)
		if err != nil {
			return err
		}
		fmt.Println(txnComment + updateStock + " successful")

		affected, err := result.RowsAffected()
		if err != nil {
			return err
		}

		if affected == 0 {
			return fmt.Errorf("stock not enough, rollback")
		}

		// insert order
		insertOrder := "insert into `orders` (`id`, `book_id`, `user_id`, `quality`) values (?, ?, ?, ?)"
		if _, err := txn.Exec(insertOrder,
			orderID, bookID, userID, amount); err != nil {
			return err
		}
		fmt.Println(txnComment + insertOrder + " successful")

		// update user
		updateUser := "update `users` set `balance` = `balance` - ? where id = ?"
		if _, err := txn.Exec(updateUser,
			price.Mul(decimal.NewFromInt(int64(amount))), userID); err != nil {
			return err
		}
		fmt.Println(txnComment + updateUser + " successful")

		return nil
	})
}

func buyOptimistic(db *sql.DB, goroutineID, orderID, bookID, userID, amount int) {
	txnComment := fmt.Sprintf("/* txn %d */ ", goroutineID)
	if goroutineID != 1 {
		txnComment = "\t" + txnComment
	}

	fmt.Printf("\nuser %d try to buy %d books(id: %d)\n", userID, amount, bookID)

	runTxn(db, true, retryTimes, func(txn *util.TiDBSqlTx) error {
		time.Sleep(time.Second)

		// read the price and stock of book
		selectBookForUpdate := "select `price`, `stock` from books where id = ? for update"
		bookRows, err := txn.Query(selectBookForUpdate, bookID)
		if err != nil {
			return err
		}
		fmt.Println(txnComment + selectBookForUpdate + " successful")
		defer bookRows.Close()

		price, stock := decimal.NewFromInt(0), 0
		if bookRows.Next() {
			err = bookRows.Scan(&price, &stock)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("book ID not exist")
		}
		bookRows.Close()

		if stock < amount {
			return fmt.Errorf("book not enough")
		}

		// update book
		updateStock := "update `books` set stock = stock - ? where id = ? and stock - ? >= 0"
		result, err := txn.Exec(updateStock, amount, bookID, amount)
		if err != nil {
			return err
		}
		fmt.Println(txnComment + updateStock + " successful")

		affected, err := result.RowsAffected()
		if err != nil {
			return err
		}

		if affected == 0 {
			return fmt.Errorf("stock not enough, rollback")
		}

		// insert order
		insertOrder := "insert into `orders` (`id`, `book_id`, `user_id`, `quality`) values (?, ?, ?, ?)"
		if _, err := txn.Exec(insertOrder,
			orderID, bookID, userID, amount); err != nil {
			return err
		}
		fmt.Println(txnComment + insertOrder + " successful")

		// update user
		updateUser := "update `users` set `balance` = `balance` - ? where id = ?"
		if _, err := txn.Exec(updateUser,
			price.Mul(decimal.NewFromInt(int64(amount))), userID); err != nil {
			return err
		}
		fmt.Println(txnComment + updateUser + " successful")

		return nil
	})
}

func createBook(txn *util.TiDBSqlTx, id int, title, bookType string,
	publishedAt time.Time, price decimal.Decimal, stock int) error {
	_, err := txn.ExecContext(context.Background(),
		"INSERT INTO `books` (`id`, `title`, `type`, `published_at`, `price`, `stock`) values (?, ?, ?, ?, ?, ?)",
		id, title, bookType, publishedAt, price, stock)
	return err
}

func createUser(txn *util.TiDBSqlTx, id int, nickname string, balance decimal.Decimal) error {
	_, err := txn.ExecContext(context.Background(),
		"INSERT INTO `users` (`id`, `nickname`, `balance`) VALUES (?, ?, ?)",
		id, nickname, balance)
	return err
}

func main() {
	optimistic, alice, bob := parseParams()
	openDB("mysql", "root:@tcp(127.0.0.1:4000)/bookshop?charset=utf8mb4", func(db *sql.DB) {
		prepareData(db, optimistic)
		buy(db, optimistic, alice, bob)
	})
}

func buy(db *sql.DB, optimistic bool, alice, bob int) {
	buyFunc := buyOptimistic
	if !optimistic {
		buyFunc = buyPessimistic
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		buyFunc(db, 1, 1000, 1, 1, bob)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		buyFunc(db, 2, 1001, 1, 2, alice)
	}()

	wg.Wait()
}

func openDB(driverName, dataSourceName string, runnable func(db *sql.DB)) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	runnable(db)
}

func parseParams() (optimistic bool, alice, bob int) {
	flag.BoolVar(&optimistic, "o", false, "transaction is optimistic")
	flag.IntVar(&alice, "a", 4, "Alice bought num")
	flag.IntVar(&bob, "b", 6, "Bob bought num")

	flag.Parse()

	fmt.Println(optimistic, alice, bob)

	return optimistic, alice, bob
}
