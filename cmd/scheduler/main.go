package main

import (
    "log"

    "github.com/hibiken/asynq"
)

func main() {
    scheduler := asynq.NewScheduler(
        asynq.RedisClientOpt{
            Addr: "localhost:6379",
        },
        &asynq.SchedulerOpts{},
    )

    defer scheduler.Shutdown()

    // register cron jobs here

    if err := scheduler.Run(); err != nil {
        log.Fatal(err)
    }
}