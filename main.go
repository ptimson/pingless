package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"

	ping "github.com/prometheus-community/pro-bing"
)

const (
	defaultCheckHost       = "8.8.8.8"
	defaultMaxFailures     = "10"
	defaultPingInterval    = "60s"
	defaultPingTimeout     = "3s"
	defaultRetryDelay      = "5s"
	defaultCmdWaitInterval = "2m"

	scriptPath = "./on-ping-fail.sh"
)

func main() {
	// Check script
	if err := ensureExecutable(scriptPath); err != nil {
		log.Fatalf("Issue with on-ping-fail.sh make sure you've bounded one to /app/on-ping-fail.sh: %v", err)
	}

	// Load configuration from environment
	checkHost := getEnvOrDefault("PING_HOST", defaultCheckHost)

	maxFailuresStr := getEnvOrDefault("MAX_FAILURES", defaultMaxFailures)
	maxFailures, err := strconv.Atoi(maxFailuresStr)
	if err != nil {
		log.Fatalf("Invalid MAX_FAILURES: %v", err)
	}

	pingIntervalStr := getEnvOrDefault("PING_INTERVAL", defaultPingInterval)
	pingInterval, err := time.ParseDuration(pingIntervalStr)
	if err != nil {
		log.Fatalf("Invalid PING_INTERVAL: %v", err)
	}

	pingTimeoutStr := getEnvOrDefault("PING_TIMEOUT", defaultPingTimeout)
	pingTimeout, err := time.ParseDuration(pingTimeoutStr)
	if err != nil {
		log.Fatalf("Invalid PING_TIMEOUT: %v", err)
	}

	retryDelayStr := getEnvOrDefault("RETRY_DELAY", defaultRetryDelay)
	retryDelay, err := time.ParseDuration(retryDelayStr)
	if err != nil {
		log.Fatalf("Invalid RETRY_DELAY: %v", err)
	}

	cmdWaitIntervalStr := getEnvOrDefault("CMD_WAIT_INTERVAL", defaultCmdWaitInterval)
	cmdWaitInterval, err := time.ParseDuration(cmdWaitIntervalStr)
	if err != nil {
		log.Fatalf("Invalid CMD_WAIT_INTERVAL: %v", err)
	}

	failureCount := 0
	for {
		reachable := pingHost(checkHost, pingTimeout)
		if !reachable {
			failureCount++
			fmt.Printf("Ping failed (%d/%d)\n", failureCount, maxFailures)
		}
		if !reachable && failureCount < maxFailures {
			time.Sleep(retryDelay)
			continue
		}
		if !reachable && failureCount >= maxFailures {
			fmt.Println("Host unreachable. Executing on-ping-fail.sh...")
			if err := execScriptCmd(scriptPath); err != nil {
				log.Printf("on-ping-fail.sh failed: %v\n", err)
				continue
			}
			failureCount = 0
			time.Sleep(cmdWaitInterval)
			continue
		}

		failureCount = 0
		fmt.Println("Host", checkHost, "reachable, checking again in", pingInterval)
		time.Sleep(pingInterval)
	}
}

func pingHost(host string, timeout time.Duration) bool {
	pinger, err := ping.NewPinger(host)
	if err != nil {
		log.Printf("Failed to create pinger: %v\n", err)
		return false
	}
	pinger.Count = 1
	pinger.Timeout = timeout
	pinger.SetPrivileged(true) // use raw ICMP

	err = pinger.Run()
	if err != nil {
		return false
	}
	stats := pinger.Statistics()
	return stats.PacketsRecv > 0
}

func execCmd(cmd string) error {
	// Run via bash -c so the whole one-liner works
	c := exec.Command("sh", "-c", cmd)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	err := c.Run()
	if err != nil {
		return fmt.Errorf("running exec command: %w", err)
	}

	return c.Run()
}

func execScriptCmd(scriptPath string) error {
	c := exec.Command(scriptPath)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		return fmt.Errorf("running script: %w", err)
	}

	return nil
}

func getEnvOrDefault(key, def string) string {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	return val
}

// ensureExecutable checks if the script exists and is executable.
// If it’s not executable, it tries to chmod +x it. If it doesn’t exist, returns error.
func ensureExecutable(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("script %s does not exist", path)
		}
		return fmt.Errorf("could not stat %s: %w", path, err)
	}

	// File must not be a directory
	if info.IsDir() {
		return fmt.Errorf("%s is a directory, not a script", path)
	}

	mode := info.Mode()
	// Check if "any execute" bit is set
	if mode&0111 == 0 {
		// Try to make it executable (+x for user/group/others)
		if err := os.Chmod(path, mode|0111); err != nil {
			return fmt.Errorf("could not chmod +x %s: %w", path, err)
		}
	}

	return nil
}
