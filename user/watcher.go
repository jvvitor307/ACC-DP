package user

import (
	"context"
	"time"
)

type Watcher struct {
	store    *Store
	interval time.Duration
	onChange func(User)
}

func NewWatcher(store *Store, interval time.Duration, onChange func(User)) *Watcher {
	return &Watcher{
		store:    store,
		interval: interval,
		onChange: onChange,
	}
}

func (w *Watcher) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			user, changed, err := w.store.Reload()
			if err != nil {
				continue
			}
			if changed && w.onChange != nil {
				w.onChange(user)
			}
		}
	}
}
