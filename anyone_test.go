package anyone

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type Dummy struct {
	ID int64
}

func makeNotifyChan() (chan struct{}, func()) {
	c := make(chan struct{})
	return c, func() {
		time.AfterFunc(1*time.Second, func() {
			c <- struct{}{}
		})
	}
}

func TestRunCase1(t *testing.T) {
	c, notify := makeNotifyChan()
	result, err := Run(
		context.Background(),
		[]Worker[*Dummy]{
			func() (*Dummy, error) {
				<-c
				return &Dummy{ID: 1}, nil
			},
			func() (*Dummy, error) {
				defer notify()
				return &Dummy{ID: 2}, nil
			},
		},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result to be non-nil")
	}
	if result.ID != 2 {
		t.Fatalf("unexpected result, got %v", result)
	}
}

func TestRunCase2(t *testing.T) {
	c, notify := makeNotifyChan()
	_, err := Run(
		context.Background(),
		[]Worker[*Dummy]{
			func() (*Dummy, error) {
				<-c
				return nil, errors.New("error in worker1")
			},
			func() (*Dummy, error) {
				defer notify()
				return nil, errors.New("error in worker2")
			},
		},
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "error in worker1" {
		t.Fatalf("unexpected error, got %v", err)
	}
}

func TestRunCase3(t *testing.T) {
	c, notify := makeNotifyChan()
	result, err := Run(
		context.Background(),
		[]Worker[*Dummy]{
			func() (*Dummy, error) {
				<-c
				return &Dummy{}, nil
			},
			func() (*Dummy, error) {
				defer notify()
				return nil, errors.New("error in worker")
			},
		},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result to be non-nil")
	}
}

func TestRunCase4(t *testing.T) {
	c, notify := makeNotifyChan()
	result, err := Run(
		context.Background(),
		[]Worker[*Dummy]{
			func() (*Dummy, error) {
				<-c
				return &Dummy{}, nil
			},
			func() (*Dummy, error) {
				defer notify()
				panic("panic in worker")
			},
		},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result to be non-nil")
	}
}

func TestRunCase5(t *testing.T) {
	c, notify := makeNotifyChan()
	_, err := Run(
		context.Background(),
		[]Worker[*Dummy]{
			func() (*Dummy, error) {
				<-c
				panic("panic in worker1")
			},
			func() (*Dummy, error) {
				defer notify()
				panic("panic in worker2")
			},
		},
	)
	if err == nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(err.Error(), "panic in worker1") {
		t.Fatalf("expected error to contain panic in worker1, got %v", err)
	}
}

func TestTimeoutCase1(t *testing.T) {
	_, err := Timeout(
		context.Background(),
		func() (*Dummy, error) {
			time.Sleep(2 * time.Second)
			return &Dummy{}, nil
		},
		1*time.Second,
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestTimeoutCase2(t *testing.T) {
	_, err := Timeout(
		context.Background(),
		func() (*Dummy, error) {
			return &Dummy{}, nil
		},
		1*time.Second,
	)
	if err != nil {
		t.Fatal("expected no error, got err")
	}
}

func TestWithTimeoutCase1(t *testing.T) {
	_, err := WithTimeout(
		context.Background(),
		func() (*Dummy, error) {
			time.Sleep(2 * time.Second)
			return &Dummy{}, nil
		},
		1*time.Second,
	)()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestWithTimeoutCase2(t *testing.T) {
	result, err := WithTimeout(
		context.Background(),
		func() (*Dummy, error) {
			return &Dummy{ID: 1}, nil
		},
		1*time.Second,
	)()
	if err != nil {
		t.Fatal("expected no error, got err")
	}
	if result.ID != 1 {
		t.Fatal("expected result to be 1")
	}
}

func TestWithTimeoutPanic(t *testing.T) {
	_, err := WithTimeout(
		context.Background(),
		func() (*Dummy, error) {
			panic("oops")
		},
		1*time.Second,
	)()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
