package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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
	ImagePullTime float64 `json:"ImagePullTime"`
	Containers    []containers
}

type containers struct {
	Name        string `json:"Name"`
	Snapshotter string `json:"Snapshotter"`
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

	// Create a list of each container and its snapshotter
	containerList := make([]containers, 0)
	for _, v := range t.Containers {
		newContainer := containers{
			Name:        v.Name,
			Snapshotter: v.Snapshotter,
		}

		containerList = append(containerList, newContainer)
	}
	c.Containers = containerList

	var imagePullStart time.Time = *t.PullStartedAt
	var imagePullStop time.Time = *t.PullStoppedAt
	c.ImagePullTime = imagePullStop.Sub(imagePullStart).Seconds()

	str, _ := json.Marshal(c)
	fmt.Println(string(str))
}
