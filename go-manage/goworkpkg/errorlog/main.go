package errorlog

import "fmt"

func ErrorLog(msg string) string {
	return fmt.Sprintf("Error %v", msg)
}
