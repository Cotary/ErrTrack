package e

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"gitlab.broearn.net/common-go/library/larkMessage"
	"gitlab.broearn.net/common-go/library/log"
	"gitlab.broearn.net/common-go/library/utils"
	"runtime"
	"strings"
)

func Err(err error, message ...string) error {
	str := strings.Join(message, "-")
	if err == nil {
		if str != "" {
			return errors.New(str)
		} else {
			return nil
		}
	}

	hasStack := GetStakeErr(err) != nil
	if hasStack {
		if len(message) > 0 {
			return errors.WithMessage(err, str)
		}
		return err
	}
	// The original error doesn't have a stack trace. Add a stack trace.
	if len(message) > 0 {
		return errors.Wrap(err, str)
	}
	return errors.WithStack(err)
}

type StackTracer interface {
	StackTrace() errors.StackTrace
}

func GetStakeErr(err error) error {
	for unwrapErr := err; unwrapErr != nil; {
		if _, ok := unwrapErr.(StackTracer); ok {
			return unwrapErr
			break
		}
		u, ok := unwrapErr.(interface {
			Unwrap() error
		})
		if !ok {
			break
		}
		unwrapErr = u.Unwrap()
	}
	return nil
}

func GetErrMessage(err error) string {
	stackList := make([]error, 0)
	for unwrapErr := err; unwrapErr != nil; {
		stackList = append(stackList, unwrapErr)
		u, ok := unwrapErr.(interface {
			Unwrap() error
		})
		if !ok {
			break
		}
		unwrapErr = u.Unwrap()
	}
	allLevel := len(stackList)
	str := "\n"
	for i, e := range stackList {
		str += fmt.Sprintf("[%d]:%s\n", i+1, e.Error())
		if stackErr, ok := e.(StackTracer); ok {
			str += fmt.Sprintf("\nstack:\n")
			isFirstErr := allLevel == i+1
			for si, sf := range stackErr.StackTrace() {
				if !isFirstErr && si == 0 {
					continue
				}
				pc := uintptr(sf) - 1
				fn := runtime.FuncForPC(pc)
				file, line := fn.FileLine(pc)
				str += fmt.Sprintf("%s:%d\n", file, line)
			}
			str += fmt.Sprintf("\n")
		}
	}
	return str
}

var LarkMessageCodeConfig *larkMessage.Config

func SendMessage(ctx context.Context, err error) error {
	msg := GetErrMessage(Err(err))
	log.WithContext(ctx).Error(msg)

	atErr, ok := err.(*CodeErr)
	if ok && atErr.Level > ErrorLevel { //排除小等级的报错
		return nil
	}
	env, _ := ctx.Value(utils.ENV).(string)
	if LarkMessageCodeConfig != nil && env != "local" {
		serverName, _ := ctx.Value(utils.ServerName).(string)
		LarkMessageCode := larkMessage.NewLarkMessage(LarkMessageCodeConfig)
		message := []string{
			"server_name:", serverName,
			"env:", env,
			"error:", msg,
		}
		requestID, ok := ctx.Value(utils.RequestID).(string)
		if ok {
			message = append(message, utils.RequestID+":", requestID)
		}
		LarkMessageCode.Message(utils.English, "Running Error", message)
		LarkMessageCode.Push(ctx)
	}
	return nil
}
