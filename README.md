# prom-pipe
The purpose of `prom-pipe` is to completely eliminate the barriers to publishing
quick [Prometheus](https://prometheus.io/) metrics from your shell.

Example:

    ls /tmp/foo | wc -l | prom-pipe -j foo -n tmp_files

## Intended Use Case
- You already use Prometheus for metrics.
- You want to add a new metric quickly, and:
  - it doesn't naturally fit into any service or job
  - you don't want to change any existing production code just for this task
  - the metric can be generated easily with a shell pipeline or script

For example, imagine you want to track the number of objects under an S3
prefix, in order to keep tabs on a recent bug fix. How would you normally
do that?

With `prom-pipe`, you can just do this:

    aws s3 ls --recursive s3://bucket/prefix | wc -l | prom-pipe -j myjob -n s3_objects

That's it! Set that up to run periodically using whatever pattern you like, and
enjoy your new metric.

While Prometheus and the Pushgateway are well-designed and conceptually simple,
publishing a one-off metric from a bash script is not as easy as it should be.

## Benefits
This tool addresses these obstacles:
* Generating Prometheus text format, including type/help information and labels
* Aggregating counters
* Configuration: your scripts do not need to know about a Pushgateway or anything else
* Remembering how to use the Pushgateway API
* Remembering how to get your HTTP client (e.g. `curl`) to do the right thing
* (Future) Standing up a Pushgateway

## Examples
These examples come from the [Pushgateway README](https://github.com/prometheus/pushgateway?tab=readme-ov-file#command-line).

### Basic untyped metric
Using the Pushgateway API:

    echo "some_metric 3.14" | curl --data-binary @- http://pushgateway.example.org:9091/metrics/job/some_job

Using `prom-pipe`:

    echo 3.14 | prom-pipe -j some_job -n some_metric

### More complex metrics
Using the Pushgateway API:

    cat <<EOF | curl --data-binary @- http://pushgateway.example.org:9091/metrics/job/some_job/instance/some_instance
    # TYPE some_metric counter
    some_metric{label="val1"} 42
    # TYPE another_metric gauge
    # HELP another_metric Just an example.
    another_metric 2398.283
    EOF

Using `prom-pipe`:

    echo 42 | prom-pipe -j some_job -n some_metric -t counter -l instance=some_instance,label=val1
    echo 2398.283 | prom-pipe -j some_job -n another_metric -t gauge -l instance=some_instance -h 'Just an example'

### Using `PROMPIPE_LABELS` for static labels
You can use the environment variable `PROMPIPE_LABELS` to set labels. This can help make your incantations even more succinct.

Using the example above:

    export PROMPIPE_LABELS='job=some_job,instance=some_instance'

    echo 42 | prom-pipe -j some_job -n some_metric -t counter -l label=va1
    echo 2398.283 | promp-pipe -j some_job -n another_metric -t counter -h 'Just an example'
