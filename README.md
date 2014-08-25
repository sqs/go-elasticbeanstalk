# go-elasticbeanstalk

This repository contains:

* **ebc**, a command-line tool that simplifies deployment of binaries to [AWS](https://aws.amazon.com)
[Elastic Beanstalk](http://aws.amazon.com/elasticbeanstalk/)
* a simple [elasticbeanstalk API client package](https://sourcegraph.com/github.com/sqs/go-elasticbeanstalk/symbols/go/github.com/sqs/go-elasticbeanstalk/elasticbeanstalk) written in [Go](http://golang.org)
* a sample Go web app that can be deployed to AWS Elastic Beanstalk, along with the necessary configuration to work around the lack of official Go support (see *Implementation details* below)

[**Documentation on Sourcegraph**](https://sourcegraph.com/github.com/sqs/go-elasticbeanstalk)

[![docs examples](https://sourcegraph.com/api/repos/github.com/sqs/go-elasticbeanstalk/badges/docs-examples.png)](https://sourcegraph.com/github.com/sqs/go-elasticbeanstalk)
[![Total views](https://sourcegraph.com/api/repos/github.com/sqs/go-elasticbeanstalk/counters/views.png)](https://sourcegraph.com/github.com/sqs/go-elasticbeanstalk)
[![status](https://sourcegraph.com/api/repos/github.com/sqs/go-elasticbeanstalk/badges/status.png)](https://sourcegraph.com/github.com/sqs/go-elasticbeanstalk)
[![authors](https://sourcegraph.com/api/repos/github.com/sqs/go-elasticbeanstalk/badges/authors.png)](https://sourcegraph.com/github.com/sqs/go-elasticbeanstalk)
[![dependencies](https://sourcegraph.com/api/repos/github.com/sqs/go-elasticbeanstalk/badges/dependencies.png)](https://sourcegraph.com/github.com/sqs/go-elasticbeanstalk)

## ebc command-line client for AWS Elastic Beanstalk

ebc makes it easy to build and deploy binary source bundles to AWS Elastic
Beanstalk. You still must use eb to configure and initialize applications and
environments. (If you want to deploy your whole git repository, just use the
official [eb tool](http://aws.amazon.com/code/6752709412171743).)

* Install ebc: `go get github.com/sqs/go-elasticbeanstalk/cmd/ebc`
* Install the [AWS Elastic Beanstalk eb command-line tool](http://aws.amazon.com/code/6752709412171743)

### Walkthrough

Let's deploy a [simple Go web
app](https://github.com/sqs/go-elasticbeanstalk/blob/master/webapp/server.go) to
[Elastic Beanstalk](http://aws.amazon.com/elasticbeanstalk/).

First, we need to create the [application and
environment](http://docs.aws.amazon.com/elasticbeanstalk/latest/dg/concepts.components.html).
Follow [AWS EB
documentation](http://docs.aws.amazon.com/elasticbeanstalk/latest/dg/create_deploy_nodejs.sdlc.html)
to get them set up. Once complete, you should be running the sample Node.js app
(it says "Congratulations"). We're going to deploy our Go web app *over* that sample app.

Now, make sure you've installed `ebc` into your `PATH`. Then, from the top-level
directory of this repository, run the following command:

```
ebc -dir=webapp deploy -h
```

Check the defaults for the `-app`, `-bucket`, `-env`, and `-label` flags. These
values are read from the `.elasticbeanstalk/config` file you set up using `eb
init`. They should refer to the application and environment you created
previously.

If these values look good, then run:

```
ebc -dir=webapp deploy
```

After a few seconds, you'll see a message like `Deploy initiated (took 5.22s)`.
Now, check the AWS Elastic Beanstalk dashboard to verify that a new application
is being deployed. Once it's complete, browsing to the environment's URL should
display the "Hello from Go!" text, along with some debugging info. You're done!

#### Deploying from multiple branches

The eb and ebc tools both support deploying from multiple branches. When you
switch to another branch (with `git checkout`), run `eb branch` to configure the
branch's deployment. The ebc tool reads eb's configuration for a branch, so
there are no extra steps beyond configuring eb correctly. To inspect the
configuration that ebc will use to deploy, run `ebc -dir=DIR deploy -h`.

The sample `webapp` in this repository displays the git branch used to deploy
it, so you can verify that branch deployment was successful.


## Implementation details

### Faking Go support in Elastic Beanstalk

**NOTE:** Since this section was written, Elastic Beanstalk added Docker support, which lets you run Go apps. If you use Docker, ignore this section. The `ebc` tool is still useful even if you are using Docker (or any other language, for that matter).

Because Elastic Beanstalk doesn't natively support Go, we have to use a few tricks (in the `webapp/` and `worker/` dirs):

1. In `.ebextensions/go.config`, we run a command to install Go on the server, using the
[commands](http://docs.aws.amazon.com/elasticbeanstalk/latest/dg/customize-containers-ec2.html#customize-containers-format-commands) config feature.
1. In `.ebextensions/server.config`, we trick Elastic Beanstalk into thinking that our Go app is a Node.js app and just tell it to run the command `go run server.go`.

More information can be found at the [Elastic Beanstalk docs for Node.js
apps](http://docs.aws.amazon.com/elasticbeanstalk/latest/dg/create_deploy_nodejs.sdlc.html).

## Contact

Contact [@sqs](https://twitter.com) with questions.
