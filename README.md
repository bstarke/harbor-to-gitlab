# harbor-to-gitlab

Listener for Harbor webhook that calls a Gitlab pipeline trigger to kick off a pipeline job

code expects a file in `/etc/secrets/` named the same as the harbor repository name. You can provide
as many files as you want as long as the base url maps to the correct gitlab instance. Kubernetes secrets
mapped to `/etc/secrets/` are the intended usage.

The Docker-Compose file is set up to map the `test-data` directory into `/etc/secrets` so you can 
put test files in there with names to match your harbor. The 3 data elements are needed for the 
api call to Gitlab for triggering the pipeline.

This app is useful for triggering pipelines for deployment when an outside process is pushing images.

You can build it via Cloud Native Buildpack or the Dockerfile. 

You can change the following code block and/or set the `BASE_GIT_URL` environment variable for your gilab.
The URL must have `3` instances of `%s` for the code to work - unless you make changes :)
```go
//change this or set it in Environment variable `BASE_GIT_URL` URL must have 3 instances of %s for code to work
var baseGitUrl string = "https://git.home.starkenberg.net/api/v4/projects/%s/ref/%s/trigger/pipeline?token=%s&variables[IMAGE_SHA]=%s"
```