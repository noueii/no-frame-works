package provider

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/go-jet/jet/v2/postgres"
)

func NewSentryProvider(env *EnvProvider) (*sentryhttp.Handler, error) {
	err := sentry.Init(sentry.ClientOptions{
		Environment:   env.sentryEnv,
		Dsn:           env.sentryDsn,
		EnableTracing: true,
		BeforeSendTransaction: func(event *sentry.Event, _ *sentry.EventHint) *sentry.Event {
			if event.Transaction == "" {
				return nil
			}

			if event.Transaction == "GET /health" {
				return nil
			}

			return event
		},
		TracesSampleRate: 1.0,
	})
	if err != nil {
		return nil, fmt.Errorf("sentry initialization failed: %w", err)
	}

	initializeDBTracing()

	sentryHandler := sentryhttp.New(sentryhttp.Options{
		Repanic: true,
	})

	return sentryHandler, nil
}

func initializeDBTracing() {
	postgres.SetQueryLogger(func(ctx context.Context, queryInfo postgres.QueryInfo) {
		now := time.Now()

		callerFile, callerLine, callerFunction := queryInfo.Caller()
		callerLog := fmt.Sprintf(
			"- Caller file: %s, line: %d, function: %s\n",
			callerFile,
			callerLine,
			callerFunction,
		)
		sqlStmt, _ := queryInfo.Statement.Sql()

		span := sentry.StartSpan(ctx, "db")
		span.StartTime = now.Add(-queryInfo.Duration)
		span.Description = sqlStmt

		span.SetData("Rows processed", strconv.Itoa(int(queryInfo.RowsProcessed)))
		span.SetData("Caller", callerLog)

		if queryInfo.Err != nil {
			span.SetData("Error", queryInfo.Err.Error())
		}

		span.Finish()
	})
}
