package query 

import (
    "unicode"
)

type TokenType int

const (
    TokenError TokenType = iota
    TokenBraceOpen
    TokenBraceClose
    TokenString
    TokenColon
    TokenNumber
    TokenBracketOpen
    TokenBracketClose
)

type Token struct {
    Type  TokenType
    Value string
}

func isWhitespace(char rune) bool {
    return unicode.IsSpace(char)
}

func isDelimiter(char rune) bool {
    return char == '{' || char == '}' || char == '"' || char == ':' || char == '[' || char == ']'
}

func tokenize(input string) []Token {
    var tokens []Token
    var currentToken string
    insideString := false

    for _, char := range input {
        if isWhitespace(char) && !insideString {
            continue
        }

        if isDelimiter(char) {
            if insideString {
                currentToken += string(char)
            } else {
                if currentToken != "" {
                    tokens = append(tokens, Token{Type: TokenString, Value: currentToken})
                    currentToken = ""
                }

                switch char {
                case '{':
                    tokens = append(tokens, Token{Type: TokenBraceOpen})
                case '}':
                    tokens = append(tokens, Token{Type: TokenBraceClose})
                case '"':
                    insideString = !insideString
                case ':':
                    tokens = append(tokens, Token{Type: TokenColon})
                case '[':
                    tokens = append(tokens, Token{Type: TokenBracketOpen})
                case ']':
                    tokens = append(tokens, Token{Type: TokenBracketClose})
                }
            }
        } else {
            currentToken += string(char)
        }
    }

    return tokens
}