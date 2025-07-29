package main

import (
	"changeme/backend/handler"
	"context"
	"fmt"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"regexp"
)

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) SendInfo(email, password string) {
	fmt.Println("Got email:", email)
	fmt.Println("Got password:", password)
}

func (a *App) DisplayLogInfo(message string) {
	fmt.Println(message)
}

func (a *App) ValidateEmail(email string) bool {
	re := regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

func (a *App) RunScan() {
	runtime.EventsEmit(a.ctx, "log", "Сканирование почты")
	handler.Start(func(msg string) {
		runtime.EventsEmit(a.ctx, "log", msg)
	})
}

func (a *App) CheckEnvContainsEmailData() bool {
	return handler.CheckEnvData(func(msg string) {
		runtime.EventsEmit(a.ctx, "env_read", msg)
	})
}
