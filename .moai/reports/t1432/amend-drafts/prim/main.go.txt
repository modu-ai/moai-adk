//go:build !windows

// Draft measurement instrument (t1432 amendment 0.4.0): platform primitives the heal lock rests on.
// Not part of the change. Standalone program, stdlib only.
package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

func main() {
	dir, err := os.MkdirTemp("", "prim")
	if err != nil {
		panic(err)
	}
	defer func() { _ = os.RemoveAll(dir) }()

	// P1: two separate opens of one file in one process contend on flock.
	p := filepath.Join(dir, "lock")
	a, _ := os.OpenFile(p, os.O_RDWR|os.O_CREATE, 0o600)
	b, _ := os.OpenFile(p, os.O_RDWR, 0)
	e1 := syscall.Flock(int(a.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	e2 := syscall.Flock(int(b.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	fmt.Printf("P1 first_lock_err=%v second_nb_err=%v second_is_EWOULDBLOCK=%v\n", e1, e2, errors.Is(e2, syscall.EWOULDBLOCK))
	_ = a.Close()
	e3 := syscall.Flock(int(b.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	fmt.Printf("P1 after_first_closed_second_nb_err=%v\n", e3)
	_ = b.Close()

	// P2: a symbolic link is not opened with O_NOFOLLOW.
	victim := filepath.Join(dir, "victim")
	_ = os.WriteFile(victim, []byte("VICTIM-BYTES"), 0o644)
	link := filepath.Join(dir, "link")
	_ = os.Symlink(victim, link)
	f, err := os.OpenFile(link, os.O_RDWR|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
	fmt.Printf("P2 open_symlink_nofollow_err=%v opened=%v\n", err, f != nil)
	if f != nil {
		_ = f.Close()
	}
	got, _ := os.ReadFile(victim)
	fmt.Printf("P2 victim_bytes=%q\n", got)

	// P3: a FIFO opened O_RDWR|O_NONBLOCK|O_NOFOLLOW returns at once and is not a regular file.
	fifo := filepath.Join(dir, "fifo")
	if err := syscall.Mkfifo(fifo, 0o644); err != nil {
		panic(err)
	}
	ff, err := os.OpenFile(fifo, os.O_RDWR|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
	if err != nil {
		fmt.Printf("P3 open_fifo_err=%v\n", err)
	} else {
		st, serr := ff.Stat()
		fmt.Printf("P3 open_fifo_returned=true stat_err=%v is_regular=%v mode=%v\n", serr, st.Mode().IsRegular(), st.Mode())
		_ = ff.Close()
	}

	// P4: a directory is not opened read-write.
	d := filepath.Join(dir, "adir")
	_ = os.Mkdir(d, 0o755)
	df, err := os.OpenFile(d, os.O_RDWR|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
	fmt.Printf("P4 open_dir_rdwr_err=%v opened=%v\n", err, df != nil)
	if df != nil {
		_ = df.Close()
	}

	// P5: exclusive create over a dangling symbolic link fails and creates nothing.
	dang := filepath.Join(dir, "dangling")
	_ = os.Symlink(filepath.Join(dir, "nowhere"), dang)
	cf, err := os.OpenFile(dang, os.O_RDWR|os.O_CREATE|os.O_EXCL|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0o600)
	_, terr := os.Lstat(filepath.Join(dir, "nowhere"))
	fmt.Printf("P5 create_excl_over_dangling_err=%v opened=%v target_created=%v\n", err, cf != nil, terr == nil)
	if cf != nil {
		_ = cf.Close()
	}
}
