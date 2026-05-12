package message

import (
	"math/rand"
	"reflect"
	"testing"
)

func TestMakeUser(t *testing.T) {
	var actual, expected []byte

	s := &MockScreen{}
	u := NewUserScreen(SimpleID("foo"), s)

	cfg := u.Config()
	cfg.Theme = MonoTheme // Mono
	u.SetConfig(cfg)

	m := NewAnnounceMsg("hello")

	defer u.Close()
	u.Send(m)
	u.HandleMsg(u.ConsumeOne())

	s.Read(&actual)
	expected = []byte(m.String() + Newline)
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Got: `%s`; Expected: `%s`", actual, expected)
	}
}

func TestPublicMsgBell(t *testing.T) {
	sender := NewUserScreen(SimpleID("alice"), &MockScreen{})
	defer sender.Close()

	s := &MockScreen{}
	receiver := NewUserScreen(SimpleID("bob"), s)
	cfg := receiver.Config()
	cfg.Bell = true
	cfg.Theme = MonoTheme
	receiver.SetConfig(cfg)
	defer receiver.Close()

	msg := NewPublicMsg("hello", sender)
	receiver.HandleMsg(msg)

	var actual []byte
	s.Read(&actual)
	rendered := string(actual)

	if len(rendered) == 0 {
		t.Fatal("expected rendered output but got empty string")
	}
	if rendered[len(rendered)-len(Newline)-len(Bel):len(rendered)-len(Newline)] != Bel {
		t.Errorf("expected BEL before newline in rendered public message, got: %q", rendered)
	}
}

func TestPublicMsgNoBellWhenDisabled(t *testing.T) {
	sender := NewUserScreen(SimpleID("alice"), &MockScreen{})
	defer sender.Close()

	s := &MockScreen{}
	receiver := NewUserScreen(SimpleID("bob"), s)
	cfg := receiver.Config()
	cfg.Bell = false
	cfg.Theme = MonoTheme
	receiver.SetConfig(cfg)
	defer receiver.Close()

	msg := NewPublicMsg("hello", sender)
	receiver.HandleMsg(msg)

	var actual []byte
	s.Read(&actual)
	for _, b := range actual {
		if b == '\007' {
			t.Errorf("expected no BEL when bell is disabled, but got one in: %q", actual)
			break
		}
	}
}

func TestPublicMsgNoBellForSelf(t *testing.T) {
	s := &MockScreen{}
	u := NewUserScreen(SimpleID("alice"), s)
	cfg := u.Config()
	cfg.Bell = true
	cfg.Theme = MonoTheme
	u.SetConfig(cfg)
	defer u.Close()

	msg := NewPublicMsg("my own message", u)
	u.HandleMsg(msg)

	var actual []byte
	s.Read(&actual)
	for _, b := range actual {
		if b == '\007' {
			t.Errorf("expected no BEL for own messages, but got one in: %q", actual)
			break
		}
	}
}

func TestRenderTimestamp(t *testing.T) {
	var actual, expected []byte

	// Reset seed for username color
	rand.Seed(1)
	s := &MockScreen{}
	u := NewUserScreen(SimpleID("foo"), s)

	cfg := u.Config()
	timefmt := "AA:BB"
	cfg.Theme = DefaultTheme
	cfg.Timeformat = &timefmt
	u.SetConfig(cfg)

	if got, want := cfg.Theme.Timestamp("foo"), `[38;05;245mfoo`+Reset; got != want {
		t.Errorf("Wrong timestamp formatting:\n got: %q\nwant: %q", got, want)
	}

	m := NewPublicMsg("hello", u)

	defer u.Close()
	u.Send(m)
	u.HandleMsg(u.ConsumeOne())

	s.Read(&actual)
	expected = []byte(`[38;05;245mAA:BB` + Reset + `  [[38;05;88mfoo[0m] hello` + Newline)
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Wrong screen output:\n Got: `%q`;\nWant: `%q`", actual, expected)
	}
}
