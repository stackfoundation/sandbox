package bootstrap

import "fmt"

var debug bool
var jsonOutput bool

func Debug(code string, message string, arguments ...interface{}) {
        if debug || jsonOutput {
                if (jsonOutput) {
                        fmt.Printf("{\"code\":\"%v\",\"message\":\"", code)
                        fmt.Printf(message, arguments...)
                        fmt.Printf("\"}")
                        fmt.Println()
                } else {
                        fmt.Printf(message, arguments...)
                        fmt.Println()
                }
        }
}
