# go-elasticbeanstalk

This is a demonstration of using [Amazon Web Services](https://aws.amazon.com)
[Elastic Beanstalk](http://aws.amazon.com/elasticbeanstalk/) PaaS with a
[Go](http://golang.org) web app.

## ebc command-line client for AWS Elastic Beanstalk

ebc makes it easy to build and deploy binary source bundles to AWS Elastic
Beanstalk. You still must use eb to configure and initialize applications and
environments. (If you want to deploy your whole git repository, just use the
official [eb tool](http://aws.amazon.com/code/6752709412171743).)

* Install ebc: `go get github.com/sourcegraph/go-elasticbeanstalk/cmd/ebc`
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


## Implementation details

### Faking Go support in Elastic Beanstalk

Because Elastic Beanstalk doesn't natively support Go, we have to use a few tricks (in the `webapp/` and `worker/` dirs):

1. In `.ebextensions/go.config`, we run a command to install Go on the server, using the
[commands](http://docs.aws.amazon.com/elasticbeanstalk/latest/dg/customize-containers-ec2.html#customize-containers-format-commands) config feature.
1. In `.ebextensions/server.config`, we trick Elastic Beanstalk into thinking that our Go app is a Node.js app and just tell it to run the command `go run server.go`.

More information can be found at the [Elastic Beanstalk docs for Node.js
apps](http://docs.aws.amazon.com/elasticbeanstalk/latest/dg/create_deploy_nodejs.sdlc.html).

## Contact

Contact [@sqs](https://twitter.com) with questions.
