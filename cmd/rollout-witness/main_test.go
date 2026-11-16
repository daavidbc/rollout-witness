package main

import "testing"

func TestSuccessfulStatus(t *testing.T) {
    result := isSuccessfulStatus(200)

    if !result {
        t.Errorf("Expected HTTP 200 to be successful")
    }
}
