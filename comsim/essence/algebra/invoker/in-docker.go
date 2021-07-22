package invoker

import (
	"bytes"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

var _ Function = &DockerFunction{}

type DockerFunction struct {
}

func (d *DockerFunction) Call(input FunctionInput) FunctionOutput {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	tar, err := archive.TarWithOptions("./demo-func/", &archive.TarOptions{})
	br, err := cli.ImageBuild(ctx, tar, types.ImageBuildOptions{
		Dockerfile: "Dockerfile",
		Tags:       []string{"demo-func"},
		Remove:     true,
	})
	if err != nil {
		panic(err)
	}

	defer br.Body.Close()
	_, err = io.Copy(os.Stdout, br.Body)

	// Configured hostConfig:
	// https://godoc.org/github.com/docker/docker/api/types/container#HostConfig
	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			"9666": []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: "9666",
				},
			},
		},
		RestartPolicy: container.RestartPolicy{
			Name: "no",
		},
		LogConfig: container.LogConfig{
			Type:   "json-file",
			Config: map[string]string{},
		},
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "demo-func",
		ExposedPorts: map[nat.Port]struct{}{
			"9666": struct{}{},
		},
		Tty: false,
		Healthcheck: &container.HealthConfig{
			Test:    []string{"NONE"},
			Retries: 0,
		},
	}, hostConfig, nil, nil, "")
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	//statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	//select {
	//case err := <-errCh:
	//	if err != nil {
	//		panic(err)
	//	}
	//case <-statusCh:
	//	fmt.Println(">>>>>> STATUS")
	//}

	time.Sleep(time.Second * 2)

	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		panic(err)
	}

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)

	p := new(bytes.Buffer)
	fmt.Fprintf(p, "%s", input)

	res, err := http.Post("http://localhost:9666/invoke", "plain/text", p)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	return string(body)
}

func NewDockerFunctionRegistry() *DockerFunctionRegistry {
	return &DockerFunctionRegistry{}
}

var _ FunctionRegistry = &DockerFunctionRegistry{}

type DockerFunctionRegistry struct {
}

func (d *DockerFunctionRegistry) Get(name FunctionID) (error, Function) {
	return nil, &DockerFunction{}
}

func (d *DockerFunctionRegistry) Register(name FunctionID, path string) error {
	return nil
}

func StartDockerRuntime(fun Func) {
	go func() {
		<-time.After(time.Second * 2)
		fmt.Println("Exiting container after 5s")
		os.Exit(0)
	}()

	http.HandleFunc("/invoke", func(writer http.ResponseWriter, request *http.Request) {
		in, _ := ioutil.ReadAll(request.Body)
		ou := fun(in)
		switch t := ou.(type) {
		case string:
			fmt.Fprint(writer, t)
		default:
			fmt.Fprintf(writer, "StartDockerRuntime: Unexpected function output type '%s', expects string", t)
		}
	})

	http.ListenAndServe(":9666", http.DefaultServeMux)
}
