package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	projectClient "github.com/rancher/types/client/project/v3"
	"github.com/urfave/cli"
)

func RunCommand() cli.Command {
	return cli.Command{
		Name:   "run",
		Usage:  "Run a new workload on your rancher server",
		Action: serviceRun,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "allow-privilege-escalation",
				Usage: "",
			},
			cli.StringSliceFlag{
				Name:  "annotation",
				Usage: "Configure annotations (key/value metadata) for the container.",
			},
			cli.StringSliceFlag{
				Name: "clusterport",
				Usage: "Cluster port to publish. If one port is provided the container " +
					"port will be mapped to the source, or specify both ports " +
					"in the format <port>:<port>",
			},
			cli.StringFlag{
				Name:  "command",
				Usage: "Command to run the container",
			},
			cli.StringFlag{
				Name:  "cpu",
				Usage: "CPU limit in milli CPUs",
			},
			cli.StringFlag{
				Name:  "cpu-reserve",
				Usage: "CPU to reserve in milli CPUs",
			},
			cli.StringFlag{
				Name:  "dns-policy",
				Usage: "Options are `ClusterFirst` or `ClusterFirstWithHostNet`",
				Value: "ClusterFirst",
			},

			cli.StringFlag{
				Name:  "entrypoint",
				Usage: "Overwrite the default ENTRYPOINT of the Image",
			},
			cli.StringSliceFlag{
				Name:  "env,e",
				Usage: "Set one or more environment variable in the form of key=value",
			},
			cli.Int64Flag{
				Name:  "fsg",
				Usage: "Filesystem group",
			},
			cli.StringFlag{
				Name:  "hostname",
				Usage: "Container host name",
			},
			cli.StringSliceFlag{
				Name: "hostport",
				Usage: "Hostport to publish. Accepts a <port>:<port> or accepts " +
					"<hostip>:<hostport>:<port> mapping",
			},
			cli.StringSliceFlag{
				Name:  "host-alias",
				Usage: "Host alias in the form [IP=HOST]",
			},
			cli.BoolFlag{
				Name:  "host-network",
				Usage: "Use Host's Network Namespace",
			},
			cli.StringFlag{
				Name:  "image",
				Usage: "Image to run",
			},
			cli.StringFlag{
				Name: "image-pull-policy",
				Usage: "Pull image on container start. Options are 'Always', " +
					"'IfNotPresent' and 'Never'",
				Value: "Always",
			},
			cli.BoolFlag{
				Name:  "interactive, i",
				Usage: "Keep STDIN open even if not attached",
			},
			cli.BoolFlag{
				Name:  "ipc",
				Usage: "Use Host's IPC Namespace",
			},
			cli.StringSliceFlag{
				Name: "label,l",
				Usage: "key/value pairs used to annotate containers and make scheduling " +
					"decisions.",
			},
			cli.StringSliceFlag{
				Name:  "lbport",
				Usage: "Load balancer port mapping in the format <port>:<port>",
			},
			cli.StringFlag{
				Name:  "memory, m",
				Usage: "Memory limit in MiB",
			},
			cli.StringFlag{
				Name:  "memory-reserve",
				Usage: "Memory to reserve in MiB",
			},
			cli.StringFlag{
				Name:  "namespace",
				Usage: "Namespace to deploy into",
			},
			cli.StringSliceFlag{
				Name: "nodeport",
				Usage: "Nodeport to publish, if one port is provided it will be " +
					"mapped to a random port or both ports can be specified in the " +
					"format <port>:<port>",
			},
			cli.StringFlag{
				Name:  "nvidia-gpu",
				Usage: "Nvidia GPU limit in milli GPUs",
			},
			cli.StringFlag{
				Name:  "nvidia-gpu-reserve",
				Usage: "Nvidia GPU to reserve in milli GPUs",
			},
			cli.BoolFlag{
				Name:  "pid",
				Usage: "Use Host's PID Namespace",
			},
			cli.BoolFlag{
				Name:  "privileged",
				Usage: "Give extended privileges to this container",
			},
			cli.BoolFlag{
				Name:  "read-only",
				Usage: "Mount the container's root filesystem as read only",
			},
			cli.BoolFlag{
				Name:  "run-as-non-root",
				Usage: "Run as Non Root",
			},
			//cli.StringFlag{
			//	Name:  "restart-policy",
			//	Usage: "Always, Never or OnFailure",
			//},
			cli.Int64Flag{
				Name:  "scale",
				Usage: "Number of containers to run",
				Value: 1,
			},
			cli.Int64Flag{
				Name:  "stop-timeout",
				Usage: "Time in seconds before container is forcefully stopped",
			},
			cli.StringFlag{
				Name:  "subdomain",
				Usage: "Container subdomain",
			},
			cli.BoolFlag{
				Name:  "tty, t",
				Usage: "Allocate a pseudo-TTY",
			},
			cli.Int64Flag{
				Name:  "user, u",
				Usage: "User ID, must be an int",
			},
			//cli.StringSliceFlag{
			//	Name:  "volume, v",
			//	Usage: "Bind mount a volume",
			//},
			cli.StringFlag{
				Name:  "workdir, w",
				Usage: "Working directory inside the container",
			},
		},
	}
}

func serviceRun(ctx *cli.Context) error {
	if ctx.String("namespace") == "" {
		return errors.New("namespace is required")
	}
	c, err := GetClient(ctx)
	if err != nil {
		return err
	}

	ports, err := processPorts(ctx)
	if nil != err {
		return err
	}

	var command []string
	if ctx.String("command") != "" {
		command = SplitStringOnSpace(ctx.String("command"))
	}

	var entrypoint []string
	if ctx.String("entrypoint") != "" {
		entrypoint = SplitStringOnSpace(ctx.String("entrypoint"))
	}

	container := projectClient.Container{
		AllowPrivilegeEscalation: boolPointer(ctx.Bool("allow-privilege-escalation")),
		Command:                  command,
		Entrypoint:               entrypoint,
		Environment:              KeyValuetoMap(ctx.StringSlice("env")),
		Image:                    ctx.String("image"),
		ImagePullPolicy:          ctx.String("image-pull-policy"),
		Name:                     ctx.Args().First(),
		Ports:                    ports,
		Privileged:               boolPointer(ctx.Bool("privileged")),
		ReadOnly:                 boolPointer(ctx.Bool("read-only")),
		Resources:                processResources(ctx),
		RunAsNonRoot:             boolPointer(ctx.Bool("run-as-non-root")),
		Stdin:                    ctx.Bool("interactive"),
		TTY:                      ctx.Bool("tty"),
		Uid:                      int64Pointer(ctx.Int64("user")),
		WorkingDir:               ctx.String("workdir"),
	}

	workload := projectClient.Workload{
		Annotations:                   KeyValuetoMap(ctx.StringSlice("annotation")),
		Containers:                    []projectClient.Container{container},
		DNSPolicy:                     ctx.String("dns-policy"),
		Fsgid:                         int64Pointer(ctx.Int64("fsg")),
		HostAliases:                   processHostAliases(ctx.StringSlice("host-alias")),
		HostIPC:                       ctx.Bool("ipc"),
		HostPID:                       ctx.Bool("pid"),
		Hostname:                      ctx.String("hostname"),
		HostNetwork:                   ctx.Bool("host-network"),
		Labels:                        KeyValuetoMap(ctx.StringSlice("label")),
		Name:                          ctx.Args().First(),
		NamespaceId:                   ctx.String("namespace"),
		RestartPolicy:                 ctx.String("restart-policy"),
		Scale:                         int64Pointer(ctx.Int64("scale")),
		Subdomain:                     ctx.String("subdomain"),
		TerminationGracePeriodSeconds: int64Pointer(ctx.Int64("stop-timeout")),
	}

	_, err = c.ProjectClient.Workload.Create(&workload)
	if nil != err {
		return err
	}
	return nil
}

func processHostAliases(s []string) []projectClient.HostAlias {
	var ha []projectClient.HostAlias
	holderMap := make(map[string][]string)
	for _, value := range s {
		splits := strings.Split(value, "=")
		if len(splits) != 2 {
			continue
		}
		holderMap[splits[0]] = append(holderMap[splits[0]], splits[1])
	}

	for ip, hosts := range holderMap {
		ha = append(ha, projectClient.HostAlias{
			IP:        ip,
			Hostnames: hosts,
		})
	}
	return ha
}

func processPorts(ctx *cli.Context) ([]projectClient.ContainerPort, error) {
	var ports []projectClient.ContainerPort
	ports, err := processMappedPorts(ports, ctx.StringSlice("nodeport"), "NodePort")
	if nil != err {
		return ports, err
	}
	ports, err = processMappedPorts(ports, ctx.StringSlice("clusterport"), "ClusterIP")
	if nil != err {
		return ports, err
	}
	ports, err = processMappedPorts(ports, ctx.StringSlice("lbport"), "LoadBalancer")
	if nil != err {
		return ports, err
	}
	// Host ports can also have an IP in them so need to process differently
	ports, err = processHostPorts(ports, ctx.StringSlice("hostport"))
	if nil != err {
		return ports, err
	}
	fmt.Println(ports)
	return ports, nil
}

func processMappedPorts(
	ports []projectClient.ContainerPort,
	portStrings []string,
	kind string,
) ([]projectClient.ContainerPort, error) {
	for _, portMapping := range portStrings {
		np := projectClient.ContainerPort{
			Kind:     kind,
			Protocol: "TCP",
		}
		pieces, err := parseStringPortMapping(portMapping)
		if nil != err {
			return ports, err
		}
		if len(pieces) == 1 {
			np.ContainerPort = pieces[0]
		} else if len(pieces) == 2 {
			np.ContainerPort = pieces[1]
			np.SourcePort = pieces[0]
		} else if len(pieces) == 1 && kind == "ClusterIP" {
			np.SourcePort = pieces[0]
		}
		ports = append(ports, np)
	}
	fmt.Println(ports)
	return ports, nil
}

// processHostPorts processes the flag hostport which can pass in hostip:hostport:port
func processHostPorts(
	ports []projectClient.ContainerPort,
	portStrings []string,
) ([]projectClient.ContainerPort, error) {
	for _, ps := range portStrings {
		np := projectClient.ContainerPort{
			Kind:     "HostPort",
			Protocol: "TCP",
		}
		pieces := strings.Split(ps, ":")
		if len(pieces) == 2 {
			portInts, err := parseStringPortMapping(ps)
			if nil != err {
				return ports, err
			}
			np.ContainerPort = portInts[1]
			np.SourcePort = portInts[0]
		} else if len(pieces) == 3 {
			np.HostIp = pieces[0]
			portInts, err := parseStringPortMapping(pieces[1] + ":" + pieces[2])
			if nil != err {
				return ports, err
			}
			np.ContainerPort = portInts[1]
			np.SourcePort = portInts[0]
		} else {
			return ports, fmt.Errorf("invalid port mapping: %v", ps)
		}
		ports = append(ports, np)
	}
	return ports, nil
}

// pass a port mapping string 'port:port' or 'port' and returns a slice of *int64 ports
func parseStringPortMapping(s string) ([]*int64, error) {
	var ports []*int64
	pieces := strings.Split(s, ":")
	for _, piece := range pieces {
		portInt, err := strconv.Atoi(piece)
		if nil != err {
			return ports, err
		}
		ports = append(ports, int64Pointer(int64(portInt)))
	}
	return ports, nil
}

func processResources(ctx *cli.Context) *projectClient.Resources {
	var resources *projectClient.Resources
	if ctx.String("cpu") != "" || ctx.String("cpu-reserve") != "" {
		resources.CPU = &projectClient.ResourceRequest{
			Limit:   ctx.String("cpu"),
			Request: ctx.String("cpu-reserve"),
		}
	}
	if ctx.String("memory") != "" || ctx.String("memory-reserve") != "" {
		resources.Memory = &projectClient.ResourceRequest{
			Limit:   ctx.String("memory"),
			Request: ctx.String("memory-reserve"),
		}
	}
	if ctx.String("nvidia-gpu") != "" || ctx.String("nvidia-gpu-reserve") != "" {
		resources.Memory = &projectClient.ResourceRequest{
			Limit:   ctx.String("nvidia-gpu"),
			Request: ctx.String("nvidia-gpu-reserve"),
		}
	}

	return resources
}
