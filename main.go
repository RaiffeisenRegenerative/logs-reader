package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/crypto/ssh"
)

type LogEntry struct {
	Level     string `json:"level"`
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
	Raw       string
	Source    string
	Time      time.Time
}

func loadPrivateKey(path string) (ssh.AuthMethod, error) {
	usr, _ := user.Current()
	if strings.HasPrefix(path, "~/") {
		path = filepath.Join(usr.HomeDir, path[2:])
	}
	key, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}
	return ssh.PublicKeys(signer), nil
}

func fetchLog(host string, port int, path string, config *ssh.ClientConfig) ([]string, error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	var out bytes.Buffer
	session.Stdout = &out

	// Automatically use zcat for .gz files
	cmd := fmt.Sprintf("case %s in *.gz) zcat %s ;; *) cat %s ;; esac", path, path, path)
	if err := session.Run(cmd); err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	return lines, nil
}

func listRemoteLogFiles(dir string, pattern string, config *ssh.ClientConfig, host string, port int) ([]string, error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	var out bytes.Buffer
	session.Stdout = &out

	cmd := fmt.Sprintf("find %s -type f -name '%s'", dir, pattern)
	if err := session.Run(cmd); err != nil {
		return nil, err
	}

	files := strings.Split(strings.TrimSpace(out.String()), "\n")
	return files, nil
}

func parseLogLine(line string, source string) (LogEntry, error) {
	var e LogEntry
	err := json.Unmarshal([]byte(line), &e)
	if err != nil {
		return LogEntry{}, err
	}
	e.Raw = line
	e.Source = source
	e.Time, _ = time.Parse("2006-01-02 15:04:05", e.Timestamp)
	return e, nil
}

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	auth, err := loadPrivateKey(cfg.SSHKey)
	if err != nil {
		fmt.Println("Failed to load SSH key:", err)
		os.Exit(1)
	}

	sshConfig := &ssh.ClientConfig{
		User:            cfg.SSHUser,
		Auth:            []ssh.AuthMethod{auth},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	var allLogs []LogEntry

	emailLines, err := fetchLog(cfg.SSHHost, cfg.EmailPort, cfg.EmailLogPath, sshConfig)
	if err == nil {
		for _, line := range emailLines {
			if entry, err := parseLogLine(line, "email"); err == nil {
				allLogs = append(allLogs, entry)
			}
		}
	}

	logFiles, err := listRemoteLogFiles("~/logs", "error*.log*", sshConfig, cfg.SSHHost, cfg.NextjsPort)
	fmt.Println("Listing remote log files:", logFiles)
	if err != nil {
		fmt.Println("Failed to list remote logs:", err)
	} else {
		for _, file := range logFiles {
			lines, err := fetchLog(cfg.SSHHost, cfg.NextjsPort, file, sshConfig)
			if err != nil {
				continue
			}
			for _, line := range lines {
				if entry, err := parseLogLine(line, "nextjs"); err == nil {
					allLogs = append(allLogs, entry)
				}
			}
		}
	}

	sort.Slice(allLogs, func(i, j int) bool {
		return allLogs[i].Time.After(allLogs[j].Time)
	})

	items := make([]list.Item, len(allLogs))
	for i, log := range allLogs {
		items[i] = logItem{entry: log}
	}

	keys := newKeyMap()
	delegate := newCustomDelegate(keys)

	l := list.New(items, delegate, 0, 0)
	l.Title = "Logs"

	m := model{
		list:         l,
		keys:         keys,
		activeOrigin: "",
		activeLevel:  "",
		allItems:     items,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running Bubble Tea UI:", err)
		os.Exit(1)
	}
}
