package crypto

import "fmt"

// Written by chatgpt, This is currently not used, but will be for
// our own encryption/decryption algorithm later.
// all we need to know is, give it ascii value and it'll return char,
// give it char and it'll give ascii value. neat!
//
//lint:ignore U1000 For future use
func asciiConverter(input interface{}) (output interface{}, err error) { // nolint:U1000
	switch v := input.(type) {
	case string:
		if len(v) != 1 {
			return nil, fmt.Errorf("input must be a single character")
		}
		return int(v[0]), nil // Convert the character to its ASCII value

	case int:
		if v < 0 || v > 255 { // Allowing values from 0 to 255
			return nil, fmt.Errorf("input must be an ASCII value (0-255)")
		}
		return string(byte(v)), nil // Convert the ASCII value to character

	default:
		return nil, fmt.Errorf("input must be a string (for character) or int (for ASCII value)")
	}
}
