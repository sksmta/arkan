package query 

import (
	"fmt"
)

func unescapeString(input string) (string, error) {
    var output string
    var escape bool

    for _, char := range input {
        if escape {
            switch char {
            case 'n':
                output += "\n"
            case 't':
                output += "\t"
            case '\\':
                output += "\\"
            case '"':
                output += "\""
            default:
                return "", fmt.Errorf("invalid escape sequence: \\%c", char)
            }
            escape = false
        } else if char == '\\' {
            escape = true
        } else {
            output += string(char)
        }
    }

    return output, nil
}