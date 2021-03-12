package main

import (
	"context"
	"control-mitsubishi-plc-r-kube/cmd"
	"control-mitsubishi-plc-r-kube/config"
	"control-mitsubishi-plc-r-kube/pkg"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	errCh := make(chan error, 1)
	ctx, cancel := context.WithCancel(context.Background())
	cfg, err := config.New()
	if err != nil {
		errCh <- err
	}
	defer cancel()
	go listen(ctx, 2*time.Second, errCh, cfg)
	s := pkg.New(ctx, cfg)
	go s.Start(errCh)

	quitC := make(chan os.Signal, 1)
	signal.Notify(quitC, syscall.SIGTERM, os.Interrupt)

	select {
	case err := <-errCh:
		cancel()
		panic(err)
	case <-quitC:
		s.Shutdown(ctx)
		cancel()
	}
}

func listen(ctx context.Context, interval time.Duration, errCh chan error, cfg *config.Config) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	err := cmd.ReadCombPlc(ctx, cfg)
	if err != nil {
		errCh <- err
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := cmd.ReadCombPlc(ctx, cfg)
			if err != nil {
				errCh <- err
			}
		}
	}
}
