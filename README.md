# go-elasticbeanstalk

This is a demonstration of using [Amazon Web Services](https://aws.amazon.com)
[Elastic Beanstalk](http://aws.amazon.com/elasticbeanstalk/) PaaS with a
[Go](http://golang.org) web app.

To set up and run the app, run `eb init`, follow the prompts, and then run `git
aws.push`.

Because Elastic Beanstalk doesn't natively support Go, we have to use a few tricks:

1. In `.ebextensions/go.config`, we run a command to install Go on the server, using the
[commands](http://docs.aws.amazon.com/elasticbeanstalk/latest/dg/customize-containers-ec2.html#customize-containers-format-commands) config feature.
1. In `.ebextensions/server.config`, we trick Elastic Beanstalk into thinking that our Go app is a Node.js app and just tell it to run the command `go run server.go`.

More information can be found at the [Elastic Beanstalk docs for Node.js
apps](http://docs.aws.amazon.com/elasticbeanstalk/latest/dg/create_deploy_nodejs.sdlc.html).

## Contact

Contact [@sqs](https://twitter.com) with questions.
