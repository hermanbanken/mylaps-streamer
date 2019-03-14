package main

import "testing"

func TestDigitsShort(t *testing.T) {
	if digits("36.730") != "367" {
		t.Errorf("Got %s, expected 367", digits("36.730"))
	}
}

func TestDigitsTooSlow(t *testing.T) {
	if digits("56.730") != "___" {
		t.Errorf("Got %s, expected ___", digits("56.730"))
	}
}

func TestDigitsTooFast(t *testing.T) {
	if digits("16.730") != "___" {
		t.Errorf("Got %s, expected ___", digits("16.730"))
	}
}

func TestDigitsLong(t *testing.T) {
	if digits("1:09.395") != "___" {
		t.Errorf("Got %s, expected ___", digits("1:09.395"))
	}
}

func TestDigitsEmpty(t *testing.T) {
	if digits("") != "___" {
		t.Errorf("Got %s, expected ___", digits(""))
	}
}
