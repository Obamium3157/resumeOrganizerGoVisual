package main

import (
	"context"
	"fmt"
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
