package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/urfave/cli"
)

var versionNumber = "0.0.0" // version number set at compile time

func main() {
	app := cli.NewApp()
	app.Name = "drone-kube"
	app.Usage = "Kubernetes Deployment plugin for Drone"
	app.Action = run
	app.Version = versionNumber
	app.Flags = makeCliFlagArray()

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("Execution failed: %v", err)
	}
}

func run(c *cli.Context) error {
	// kubernetes token
	if c.String("env-file") != "" {
		_ = godotenv.Load(c.String("env-file"))
	}

	plugin := plugin{
		Repo: repo{
			Owner: c.String("repo.owner"),
			Name:  c.String("repo.name"),
		},
		Build: build{
			Tag:     c.String("build.tag"),
			Number:  c.Int("build.number"),
			Event:   c.String("build.event"),
			Status:  c.String("build.status"),
			Commit:  c.String("commit.sha"),
			Ref:     c.String("commit.ref"),
			Branch:  c.String("commit.branch"),
			Author:  c.String("commit.author"),
			Link:    c.String("build.link"),
			Started: c.Int64("build.started"),
			Created: c.Int64("build.created"),
		},
		Job: job{
			Started: c.Int64("job.started"),
		},
		Config: config{
			Kubeconfig: c.String("kubeconfig"),
			ClientCert: c.String("clientcert"),
			ClientKey:  c.String("clientkey"),
			Token:      c.String("token"),
			Server:     c.String("server"),
			Ca:         c.String("ca"),
			Namespace:  c.String("namespace"),
			Template:   c.String("template"),
		},
	}

	return plugin.exec()
}

func makeCliFlagArray() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:   "kubeconfig",
			Usage:  "Kubernetes config file",
			EnvVar: "KUBE_CONFIG",
		},
		cli.StringFlag{
			Name:   "clientcert",
			Usage:  "PEM-encoded client certificate file",
			EnvVar: "KUBE_CLIENT_CERT",
		},
		cli.StringFlag{
			Name:   "clientkey",
			Usage:  "PEM-encoded client certificate key file",
			EnvVar: "KUBE_CLIENT_KEY",
		},
		cli.StringFlag{
			Name:   "token",
			Usage:  "Kubernetes bearer token",
			EnvVar: "KUBE_TOKEN,PLUGIN_TOKEN",
		},
		cli.StringFlag{
			Name:   "ca",
			Usage:  "Certificate Authority file encoded into base64: e.g: run: `cat ca.pem | base64` to get this value",
			EnvVar: "KUBE_CA,PLUGIN_CA",
		},
		cli.StringFlag{
			Name:   "server",
			Usage:  "Server url: e.g: https://mykubernetes:6433",
			EnvVar: "KUBE_SERVER,PLUGIN_SERVER",
		},
		cli.StringFlag{
			Name:   "namespace",
			Usage:  "namespace to use: 'default' is the default :-)",
			EnvVar: "KUBE_NAMESPACE,PLUGIN_NAMESPACE",
		},
		cli.StringFlag{
			Name:   "template",
			Usage:  "template file to use for deployment: mydeployment.yaml :-)",
			EnvVar: "KUBE_TEMPLATE,PLUGIN_TEMPLATE",
		},
		cli.StringFlag{
			Name:   "repo.owner",
			Usage:  "repository owner",
			EnvVar: "DRONE_REPO_OWNER",
		},
		cli.StringFlag{
			Name:   "repo.name",
			Usage:  "repository name",
			EnvVar: "DRONE_REPO_NAME",
		},
		cli.StringFlag{
			Name:   "commit.sha",
			Usage:  "git commit sha",
			EnvVar: "DRONE_COMMIT_SHA",
		},
		cli.StringFlag{
			Name:   "commit.ref",
			Value:  "refs/heads/master",
			Usage:  "git commit ref",
			EnvVar: "DRONE_COMMIT_REF",
		},
		cli.StringFlag{
			Name:   "commit.branch",
			Value:  "master",
			Usage:  "git commit branch",
			EnvVar: "DRONE_COMMIT_BRANCH",
		},
		cli.StringFlag{
			Name:   "commit.author",
			Usage:  "git author name",
			EnvVar: "DRONE_COMMIT_AUTHOR",
		},
		cli.StringFlag{
			Name:   "build.event",
			Value:  "push",
			Usage:  "build event",
			EnvVar: "DRONE_BUILD_EVENT",
		},
		cli.IntFlag{
			Name:   "build.number",
			Usage:  "build number",
			EnvVar: "DRONE_BUILD_NUMBER",
		},
		cli.StringFlag{
			Name:   "build.status",
			Usage:  "build status",
			Value:  "success",
			EnvVar: "DRONE_BUILD_STATUS",
		},
		cli.StringFlag{
			Name:   "build.link",
			Usage:  "build link",
			EnvVar: "DRONE_BUILD_LINK",
		},
		cli.Int64Flag{
			Name:   "build.started",
			Usage:  "build started",
			EnvVar: "DRONE_BUILD_STARTED",
		},
		cli.Int64Flag{
			Name:   "build.created",
			Usage:  "build created",
			EnvVar: "DRONE_BUILD_CREATED",
		},
		cli.StringFlag{
			Name:   "build.tag",
			Usage:  "build tag",
			EnvVar: "DRONE_TAG",
		},
	}
}
