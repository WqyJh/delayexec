package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
)

const (
	TimeFormat = "2006-01-02 15:04:05"
)

var (
	argCfg      string
	argLog      string
	argDuration time.Duration
	argTime     string
	argCron     string
	argWorkDir  string
	argCancel   bool
)

var (
	cfgPath string
	workDir string
)

func TimeToCron(due time.Time) string {
	return fmt.Sprintf("%d %d %d %d *",
		due.Minute(), due.Hour(), due.Day(), due.Month())
}

func GetCrontab() string {
	var output bytes.Buffer
	cmd := exec.Command("crontab", "-l")
	cmd.Stdout = &output
	_ = cmd.Run()
	return output.String()
}

func SetCrontab(crontab string) {
	cmd := exec.Command("crontab", "-")
	cmd.Stdin = strings.NewReader(crontab)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Fatalf("failed to set crontab: %s", err.Error())
	}
}

func AppendCrontab(cron string) {
	crontab := GetCrontab()
	if strings.Contains(crontab, cron) {
		// dedup
		return
	}
	if crontab == "" {
		crontab = cron
	} else {
		crontab = fmt.Sprintf("%s\n%s", crontab, cron)
	}
	SetCrontab(crontab)
}

func RemoveCrontab(cron string) {
	crontab := GetCrontab()
	if strings.Contains(crontab, cron) {
		crontab = strings.ReplaceAll(crontab, cron, "")
		SetCrontab(crontab)
		return
	}
}

func isDirectory(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}

	return fileInfo.IsDir()
}

func GetOneshotScriptPath(t time.Time, cmds []string) string {
	directory := path.Join(cfgPath, cmds[0])

	if !isDirectory(directory) {
		err := os.MkdirAll(directory, 0755)
		if err != nil {
			log.Fatalf("create directory '%s' error: %s", directory, err.Error())
		}
	}

	script := fmt.Sprintf("%d.sh", t.Unix())
	script = path.Join(directory, script)
	log.Printf("scriptPath: %s", script)
	return script
}

func GenerateOneshotScript(t time.Time, cmds []string) string {
	command := strings.Join(cmds, " ")
	var cancels []string

	delayexec, err := os.Executable()
	if err != nil {
		log.Fatalf("executable path of '%s' not found: %s", os.Args[0], err.Error())
	}

	cancels = append(cancels, delayexec)
	cancels = append(cancels, "--cancel")
	cancels = append(cancels, os.Args[1:]...)
	for i := 0; i < len(cancels); i++ {
		if cancels[i] == "-d" || cancels[i] == "--duration" {
			cancels[i] = "-t"
			cancels[i+1] = t.Format(TimeFormat)
		}
		re := regexp.MustCompile(`\s+`)
		if re.Match([]byte(cancels[i])) {
			cancels[i] = fmt.Sprintf(`"%s"`, cancels[i])
		}
	}
	var cancelCmd = strings.Join(cancels, " ")
	return fmt.Sprintf(`#!/bin/sh
cd %s
%s >>%s 2>&1
%s >>%s 2>&1
`, workDir, command, argLog, cancelCmd, argLog)
}

func SetCron(cron string, cmds []string) {
	cmd := strings.Join(cmds, " ")
	crontab := fmt.Sprintf("%s %s", cron, cmd)
	if argCancel {
		RemoveCrontab(crontab)
	} else {
		AppendCrontab(crontab)
	}
}

func SetOneshot(due time.Time, cmds []string) {
	path := GetOneshotScriptPath(due, cmds)

	if argCancel {
		os.Remove(path)
	} else {
		script := GenerateOneshotScript(due, cmds)
		err := ioutil.WriteFile(path, []byte(script), 0755)
		if err != nil {
			log.Fatalf("write file '%s' failed: %s", path, err.Error())
		}
	}

	cron := TimeToCron(due)
	crontab := fmt.Sprintf("%s %s\n", cron, path)
	if argCancel {
		RemoveCrontab(crontab)
	} else {
		AppendCrontab(crontab)
	}
}

func cmdRun(cmd *cobra.Command, args []string) {
	if argCfg == "" {
		argCfg = ".delayexec"
	}

	homePath, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("unable to get user home dir: %s", err.Error())
	}
	cfgPath = path.Join(homePath, argCfg)

	workDir = argWorkDir
	if workDir == "" {
		workDir, err = os.Getwd()
		if err != nil {
			log.Fatalf("unable to get current working directory: %s", err.Error())
		}
	}

	if argCron != "" {
		parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
		_, err := parser.Parse(argCfg)
		if err != nil {
			log.Fatalf("failed parse cron '%s': %s", argCron, err.Error())
		}
		SetCron(argCron, args)
		return
	}

	if argTime != "" {
		t, err := time.Parse(TimeFormat, argTime)
		if err != nil {
			log.Fatalf("failed parse time '%s': %s", argTime, err.Error())
		}
		SetOneshot(t, args)
		return
	}

	if argDuration <= 0 {
		log.Fatalf("duration must be positive: %d", argDuration)
	}

	due := time.Now().Add(argDuration)
	// remove millis/nanos
	due, _ = time.Parse(TimeFormat, due.Format(TimeFormat))
	SetOneshot(due, args)
}

var rootCmd = &cobra.Command{
	Use:   "delayexec",
	Short: "delayexec is a command line tool to delay command execution.",
	Long:  `delayexec depends on crontab to delay command execution, which supports oneshot execution and repeat execution.`,
	Run:   cmdRun,
}

func init() {
	flags := rootCmd.PersistentFlags()
	flags.StringVarP(&argCfg, "path", "p", ".delayexec/", "config persist path (default is $HOME/.delayexec)")
	flags.StringVarP(&argLog, "logfile", "l", "delayexec.log", "log file path for for command output")
	flags.StringVarP(&argTime, "time", "t", "", "[oneshot] Exact time to be executed in layout "+TimeFormat)
	flags.DurationVarP(&argDuration, "duration", "d", 0, "[oneshot] Duration before command to be executed. eg: 300ms/2h45m, must be positive")
	flags.StringVarP(&argCron, "cron", "c", "", "[repeat] Crontab for command scheduling. eg: \"* 2 * * *\"")
	flags.StringVarP(&argWorkDir, "workdir", "w", "", "command work directory (default is current directory)")
	flags.BoolVar(&argCancel, "cancel", false, "cancel delayed command")
}

func main() {
	rootCmd.Execute()
}
