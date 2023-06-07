package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	v4 "github.com/aws/amazon-ecs-agent/ecs-agent/tmds/handlers/v4/state"
)

type taskConfig struct {
	Cluster       string  `json:"Cluster"`
	TaskARN       string  `json:"TaskARN"`
	Family        string  `json:"Family"`
	Revision      string  `json:"Revision"`
	TaskCpu       int64   `json:"TaskCpu"`
	TaskMemory    int64   `json:"TaskMemory"`
	ImagePullTime float64 `json:"ImagePullTime,omitempty"`
	Snapshotter   string  `json:"Snapshotter,omitempty"`
}

// A utility to convert the values to proper strings.
func int8ToStr(arr []int8) string {
	b := make([]byte, 0, len(arr))
	for _, v := range arr {
		if v == 0x00 {
			break
		}
		b = append(b, byte(v))
	}
	return string(b)
}

func findCores() int64 {
	out, _ := exec.Command("cat", "/proc/cpuinfo").Output()
	outstring := strings.TrimSpace(string(out))
	lines := strings.Split(outstring, "\n")
	var cpus int64
	cpus = 0

	for _, line := range lines {
		fields := strings.Split(line, ":")
		if len(fields) < 2 {
			continue
		}
		key := strings.TrimSpace(fields[0])

		switch key {
		case "processor":
			cpus += 1
		}
	}

	return cpus
}

func findMemory() int64 {
	out, _ := exec.Command("cat", "/proc/meminfo").Output()
	outstring := strings.TrimSpace(string(out))
	lines := strings.Split(outstring, "\n")
	var memory int64

	for _, line := range lines {
		fields := strings.Split(line, ":")
		if len(fields) < 2 {
			continue
		}
		key := strings.TrimSpace(fields[0])
		value := strings.TrimSpace(fields[1])

		switch key {
		case "MemTotal":
			valueInt := strings.Split(value, "kB")
			memoryStr := strings.TrimSpace(valueInt[0])
			m, _ := strconv.ParseInt(memoryStr, 10, 64)
			memory = m / 1024
		}
	}

	return memory
}

func readFile(filename string) string {
	var fileContents string

	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Count is a hacky way to ensure we are only getting the first line from a file
	count := 0
	for scanner.Scan() {
		fileContents = scanner.Text()
		if count == 0 {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	return fileContents
}

func areWeSoci(mountString string) string {
	// Split the string into each word, the 10th word contains the actual mount
	// paths that we are interested in. An example mount string.
	// 617 571 0:62 / / rw,relatime master:442 - overlay overlay rw,seclabel,lowerdir=/var/lib/containerd/io.containerd.snapshotter.v1.overlayfs/snapshots/1660/fs:/var/lib/containerd/io.containerd.snapshotter.v1.overlayfs/snapshots/1659/fs:/var/lib/containerd/io.containerd.snapshotter.v1.overlayfs/snapshots/1658/fs:/var/lib/containerd/io.containerd.snapshotter.v1.overlayfs/snapshots/1657/fs:/var/lib/containerd/io.containerd.snapshotter.v1.overlayfs/snapshots/1656/fs:/var/lib/containerd/io.containerd.snapshotter.v1.overlayfs/snapshots/1655/fs,upperdir=/var/lib/containerd/io.containerd.snapshotter.v1.overlayfs/snapshots/1662/fs,workdir=/var/lib/containerd/io.containerd.snapshotter.v1.overlayfs/snapshots/1662/work
	wordArray := strings.Fields(mountString)

	// The mount string contains overlyfs upperdir and lowerdir, the string for
	// lowerdir is more complex then upperdir, so we want to grab that. Is it
	// the 4th variable in the mount path. Format is:
	// rw,seclabel,lowerdir,upperdir
	mountsArray := strings.Split(wordArray[10], ",")

	// Split the Upper Dir to get the strings. Format is:
	// upperdir=/var/lib/filepath
	upperdirArray := strings.Split(mountsArray[3], "=")
	filePath := strings.Split(upperdirArray[len(upperdirArray)-1], "/")

	var snapshotter string
	snapshotter = "overlay"

	for _, directory := range filePath {
		if directory == "soci-snapshotter-grpc" {
			snapshotter = "soci"
			break
		}
	}

	return snapshotter

}

func main() {

	var c taskConfig

	metadataEndpoint := os.Getenv("ECS_CONTAINER_METADATA_URI_V4")
	if metadataEndpoint == "" {
		panic("Task Metadata Environment Variable has not been set")
	}

	// Retrieve Objects from Task Metadata Endpoint
	resp, err := http.Get(metadataEndpoint + "/task")
	if err != nil {
		print(err)
	}

	var t v4.TaskResponse
	err = json.NewDecoder(resp.Body).Decode(&t)
	if err != nil {
		panic(err)
	}

	if t.Limits == nil {
		panic("No task limits set")
	}

	c.Cluster = t.Cluster
	c.TaskARN = t.TaskARN
	c.Family = t.Family
	c.Revision = t.Revision
	c.TaskCpu = int64(*t.Limits.CPU * 1024)
	c.TaskMemory = int64(*t.Limits.Memory)
	c.Snapshotter = t.Containers[0].Snapshotter

	var imagePullStart time.Time = *t.PullStartedAt
	var imagePullStop time.Time = *t.PullStoppedAt
	c.ImagePullTime = imagePullStop.Sub(imagePullStart).Seconds()

	str, _ := json.Marshal(c)
	fmt.Println(string(str))
}
