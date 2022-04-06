# delayexec

![Build](https://github.com/WqyJh/delayexec/workflows/Tests/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/WqyJh/delayexec?style=flat-square)](https://goreportcard.com/report/WqyJh/delayexec)
[![Version](https://img.shields.io/github/release/WqyJh/delayexec.svg?style=flat-square)](https://github.com/WqyJh/delayexec/releases/latest)

Delay command execution depends on crontab.

For oneshot command, automatically calculate due time and add command to crontab, remove from crontab after the command was executed. For repeat command, just add to crontab.

Note: Make sure crond service is running.

## Install

Download pre-built binary from [Releases page](https://github.com/WqyJh/delayexec/releases).

```bash
wget -c https://github.com/WqyJh/delayexec/releases/latest/download/delayexec-linux-amd64.tar.gz -O - | tar -xz delayexec
chmod +x delayexec
sudo mv delayexec /usr/local/bin
```

Or if you have golang environment, just install with go.

```bash
go install github.com/WqyJh/delayexec@latest
```

## Usage

Execute `ls -al` after 10 minute. Duration must match golang's `time.ParseDuration` format.

```bash
delayexec -d 10m -- ls -al
```

Execute `ls` at `2022-04-05 12:00:00` in default time zone. Time format is `2006-01-02 15:04:05`.

```bash
delayexec -t '2022-04-05 12:00:00' -- ls
```

The command would be executed in current working directory, and an `delayexec.log` would be generated in it.

Use `-w` to change working directory, `-l` to change log file.

```bash
delayexec -w /tmp -l ls.log -t '2022-04-05 12:00:00' -- ls
```

## Cancel execution

If you decide to cancel the command to be executed

- with `-t` option: just add `--cancel` option such as

    ```bash
    delayexec --cancel -t '2022-04-05 12:00:00' -- ls
    ```
- with `-d` option: cannot be directly canceled.

    1. Use `crontab -l` to see the generated script.
        ```bash
        $ crontab -l
        0 2 10 4 * /home/ubuntu/.delayexec/git/1649556000.sh
        ```
    2. Use `cat` find the line with `--cancel` option.
        ```bash
        $ cat /home/ubuntu/.delayexec/git/1649556000.sh
        #!/bin/sh
        cd /home/ubuntu/docker-ssh
        git push origin master >>delayexec.log 2>&1
        /home/ubuntu/docker-ssh/delayexec --cancel -t "2022-04-10 2:00:00" -- ls -al >>delayexec.log 2>&1
        ```

    3. Execute the command with `--cancel` option.

        ```bash
        /home/ubuntu/docker-ssh/delayexec --cancel -t "2022-04-10 2:00:00" -- ls -al
        ```
