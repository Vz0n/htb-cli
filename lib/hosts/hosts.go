package hosts

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
)

const hostsFile = "/etc/hosts"

func readHostsFile(processLine func(string) (string, bool)) (string, bool, error) {
	file, err := os.Open(hostsFile)
	if err != nil {
		return "", false, err
	}
	defer file.Close()

	var buffer bytes.Buffer
	changeMade := false

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			buffer.WriteString("\n")
			continue
		}

		processedLine, changed := processLine(line)
		if changed {
			changeMade = true
		}
		buffer.WriteString(processedLine + "\n")
	}

	return buffer.String(), changeMade, scanner.Err()
}

func updateHostsFile(newContent string) error {
	isRoot := os.Getenv("USER") == "root"

	if !isRoot {
		return errors.New("you must run htb-cli as root to use this function")
	}

	file, err := os.OpenFile(hostsFile, os.O_RDWR|os.O_TRUNC, os.FileMode(0755))

	if err != nil {
		return err
	}

	// Just write to the file and close it, nothing else.
	file.WriteString(newContent)
	file.Close()

	return nil
}

func AddEntryToHosts(ip string, host string) error {
	ipFound := false
	hostAdded := false

	processLine := func(line string) (string, bool) {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			return line, false
		}

		fields := strings.Fields(trimmedLine)
		if fields[0] == ip {
			ipFound = true
			for _, field := range fields[1:] {
				if field == host {
					return line, false
				}
			}
			return line + " " + host, true
		}
		return line, false
	}

	newContent, changeMade, err := readHostsFile(processLine)
	if err != nil {
		return err
	}

	if !ipFound {
		newContent = strings.TrimSpace(newContent) + "\n" + ip + " " + host + "\n"
		hostAdded = true
	} else {
		hostAdded = changeMade
	}

	if hostAdded {
		if err := updateHostsFile(newContent); err != nil {
			return err
		}
		fmt.Println("Entry successfully updated or added.")
		return nil
	}

	fmt.Println("Entry already exists.")
	return nil
}

func RemoveEntryFromHosts(ip string, host string) error {
	hostRemoved := false

	processLine := func(line string) (string, bool) {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			return line, false
		}

		fields := strings.Fields(trimmedLine)
		if fields[0] == ip {
			var newFields []string
			newFields = append(newFields, ip)

			for _, field := range fields[1:] {
				if field != host {
					newFields = append(newFields, field)
				}
			}

			if len(newFields) > 1 {
				return strings.Join(newFields, " "), true
			}
			return "", true
		}
		return line, false
	}

	newContent, changeMade, err := readHostsFile(processLine)
	if err != nil {
		return err
	}

	if changeMade {
		newContent = strings.TrimSpace(newContent) + "\n" // Keep the end of file newline
		if err := updateHostsFile(newContent); err != nil {
			return err
		}
		fmt.Println("Entry successfully deleted.")
		return nil
	}

	if !hostRemoved {
		fmt.Println("Entry not found.")
	}
	return nil
}
