package server

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"hertz-study/pkg/app/middlewares/server/recovery"
	"hertz-study/pkg/common/config"
	"hertz-study/pkg/common/errors"
	"hertz-study/pkg/common/hlog"
	"hertz-study/pkg/route"
)

// 路由结构体
// Hertz is the core struct of hertz.
type Hertz struct {
	*route.Engine
	signalWaiter func(err chan error) error
}

// 创建一个新引擎
// New creates a hertz instance without any default config.
func New(opts ...config.Option) *Hertz {
	// 生成可选项
	options := config.NewOptions(opts)
	h := &Hertz{
		Engine: route.NewEngine(options),
	}
	return h
}

// Default creates a hertz instance with default middlewares.
func Default(opts ...config.Option) *Hertz {
	h := New(opts...)
	h.Use(recovery.Recovery())

	return h
}

// 启动函数
// Spin runs the server until catching os.Signal or error returned by h.Run().
func (h *Hertz) Spin() {
	errCh := make(chan error)
	h.initOnRunHooks(errCh)
	// 运行服务器，并监听错误
	go func() {
		errCh <- h.Run()
	}()
	// 关机信号量
	signalWaiter := waitSignal
	if h.signalWaiter != nil {
		signalWaiter = h.signalWaiter
	}

	if err := signalWaiter(errCh); err != nil {
		hlog.SystemLogger().Errorf("Receive close signal: error=%v", err)
		if err := h.Engine.Close(); err != nil {
			hlog.SystemLogger().Errorf("Close error=%v", err)
		}
		return
	}

	hlog.SystemLogger().Infof("Begin graceful shutdown, wait at most num=%d seconds...", h.GetOptions().ExitWaitTimeout/time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), h.GetOptions().ExitWaitTimeout)
	defer cancel()

	if err := h.Shutdown(ctx); err != nil {
		hlog.SystemLogger().Errorf("Shutdown error=%v", err)
	}
}

// SetCustomSignalWaiter sets the signal waiter function.
// If Default one is not met the requirement, set this function to customize.
// Hertz will exit immediately if f returns an error, otherwise it will exit gracefully.
func (h *Hertz) SetCustomSignalWaiter(f func(err chan error) error) {
	h.signalWaiter = f
}

// Default implementation for signal waiter.
// SIGTERM triggers immediately close.
// SIGHUP|SIGINT triggers graceful shutdown.
func waitSignal(errCh chan error) error {
	signalToNotify := []os.Signal{syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM}
	if signal.Ignored(syscall.SIGHUP) {
		signalToNotify = []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, signalToNotify...)
	// 开启监听
	select {
	case sig := <-signals:
		switch sig {
		case syscall.SIGTERM:
			// force exit
			return errors.NewPublic(sig.String()) // nolint
		case syscall.SIGHUP, syscall.SIGINT:
			hlog.SystemLogger().Infof("Received signal: %s\n", sig)
			// graceful shutdown
			return nil
		}
	case err := <-errCh:
		// error occurs, exit immediately
		return err
	}

	return nil
}

// 初始化运行回调函数
func (h *Hertz) initOnRunHooks(errChan chan error) {
	// add register func to runHooks
	opt := h.GetOptions()
	h.OnRun = append(h.OnRun, func(ctx context.Context) error {
		go func() {
			// delay register 1s
			time.Sleep(1 * time.Second)
			if err := opt.Registry.Register(opt.RegistryInfo); err != nil {
				hlog.SystemLogger().Errorf("Register error=%v", err)
				// pass err to errChan
				errChan <- err
			}
		}()
		return nil
	})
}
