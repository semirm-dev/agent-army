package tui

import (
	"io"
	"testing"
)

func TestSelectOne(t *testing.T) {
	p := NewFakePrompter("2")
	got, err := SelectOne(p, io.Discard, "Pick:", []string{"a", "b", "c"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "b" {
		t.Errorf("got %q, want %q", got, "b")
	}
}

func TestSelectMulti(t *testing.T) {
	p := NewFakePrompter("1,3")
	got, err := SelectMulti(p, io.Discard, "Pick:", []string{"a", "b", "c"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != "a" || got[1] != "c" {
		t.Errorf("got %v, want [a c]", got)
	}
}

func TestPromptWithDefault_Empty(t *testing.T) {
	p := NewFakePrompter("")
	got, err := PromptWithDefault(p, "Name", "default")
	if err != nil {
		t.Fatal(err)
	}
	if got != "default" {
		t.Errorf("got %q, want %q", got, "default")
	}
}

func TestPromptWithDefault_Value(t *testing.T) {
	p := NewFakePrompter("custom")
	got, err := PromptWithDefault(p, "Name", "default")
	if err != nil {
		t.Fatal(err)
	}
	if got != "custom" {
		t.Errorf("got %q, want %q", got, "custom")
	}
}

func TestSelectOneWithDefault_Enter(t *testing.T) {
	p := NewFakePrompter("")
	got, err := SelectOneWithDefault(p, io.Discard, "Pick:", []string{"a", "b"}, "b")
	if err != nil {
		t.Fatal(err)
	}
	if got != "b" {
		t.Errorf("got %q, want %q", got, "b")
	}
}

func TestSelectOneWithDefault_Number(t *testing.T) {
	p := NewFakePrompter("1")
	got, err := SelectOneWithDefault(p, io.Discard, "Pick:", []string{"a", "b"}, "b")
	if err != nil {
		t.Fatal(err)
	}
	if got != "a" {
		t.Errorf("got %q, want %q", got, "a")
	}
}

func TestSelectMultiOptional_Empty(t *testing.T) {
	p := NewFakePrompter("")
	got, err := SelectMultiOptional(p, io.Discard, "Pick:", []string{"a", "b"})
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Errorf("got %v, want nil", got)
	}
}

func TestSelectMultiOptional_Selection(t *testing.T) {
	p := NewFakePrompter("1,2")
	got, err := SelectMultiOptional(p, io.Discard, "Pick:", []string{"a", "b", "c"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Errorf("got %v, want [a b]", got)
	}
}
