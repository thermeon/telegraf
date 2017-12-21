# nsca Output Plugin

This plugin writes to [Icinga](icinga.dev.thermeon.com:5668) via tcp.

### Configuration:

# Configuration for Icinga server to send metrics to
[[outputs.Icinga]]

## The full TCP URL for your Icinga instance.
##
## Multiple urls can be specified as part of the same cluster,
## this means that only ONE of the urls will be written to each interval.
# urls = ["tcp", "icinga.dev.thermeon.com:5668", config] # TCP endpoint example
urls = ["icinga.dev.thermeon.com:5668"] # required
## The target icinga syslog will write all this information for metrics (telegraf will create it if not exists).
database = "telegraf" # required