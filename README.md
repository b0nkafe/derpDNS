# derpDNS 

The derpest dynamic DNS updater on the derp web.

## Description

derpDNS simply updates a DNS record of the type you decide. It's main purpose is to dynamically update a single A record making it point to an IP, using **OVH** API. Think of it as a clone of a <put_your_dyndns_provider_here> client for OVH customers.

derpDNS is expected to be run in a scheduled task. Below you'll find detailed instructions about how to schedule a task on macOS, Linux and ~~Windows~~. They should work even if no user is logged in.

## Usage

```
% derpDNS config_file
```

## Configuration

`derpDNS` uses a simple `json` file to define its configuration. Please check [this link](https://api.ovh.com/g934.first_step_with_api) for details on how to get your keys.

```json
{
	"record": {
		"subDomain": "subdomain",
		"zone": "domain.com",
		"recordType": "A"
	},
	"ovh": {
		"endpoint": "ovh-eu",
		"application_key": "app_key",
		"application_secret": "app_secret",
		"consumer_key": "consumer_key"
	}
}
```

This example will update `subdomain.domain.com` A record to match your current IP address. As you can check in the code, if the record doesn't exist `derpDNS` will create it.

### macOS

In macOS you should use `launchd` in order to set up the scheduled task. In order to do so, grab your `$EDITOR` and define a *LaunchDaemon* with the following XML file. Then save/copy it to `/Library/LaunchDaemons/com.b0nk.derpDNS.plist`. You can use the one included in this repo as template. Please note that you may need `root` to create/copy the file to `/Library/LaunchDaemons/`.

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
	<dict>
		<key>Label</key>
		<string>com.b0nk.derpDNS</string>
		<key>UserName</key>
		<string>your_username</string>
		<key>ProgramArguments</key>
			<array>
				<string>/path/to/derpDNS/binary</string>
				<string>/path/to/derpDNS/config</string>
			</array>
			<key>RunAtLoad</key>
			<true/>
			<key>StartInterval</key>
			<integer>3600</integer>
			<key>StandardErrorPath</key>
			<string>/tmp/derpDNS.log</string>
        </dict>
</plist>
```

All right, let's explain what we did here:

1. We set the label (name) for this LaunchDaemon to `com.b0nk.derpDNS`.
2. The LaunchDaemon will run with `my_username` permissions.
3. It will execute `/path/to/derpDNS/binary` with `/path/to/derpDNS/config` as argument.
4. We tell launchd that the program should run right after we load the `LaunchDaemon` with the key `RunAtLoad`
5. The task will run every hour (3600 secs)
6. As `derpDNS` spits everything out to STDERR, a key named `StandardErrorPath` tells launchd where STDERR will be logged for `derpDNS`. In this case, you will have a log containing every bit of output on `/tmp/derpDNS.log` 


Now, once our LaunchDaemon is set up, we should tell `launchd` to launch it:

```
% sudo launchctl load /Library/LaunchDaemons/com.b0nk.derpDNS.plist
```

No news are always good news, so `Console.App` shouldn't spit anything if everything went fine. Now you can go and check `/tmp/derpDNS.log` to see if it did something.

### Linux

It's simpler in Linux. You only have to edit `/etc/crontab` and add the following line:

```bash
0 * * * * your_username cd /path/to/derpDNS/binary && /path/to/derpDNS/binary /path/to/derpDNS/config | logger -t derpDNS
```

It should work right away.

## TODO
- [ ] Find out an easy way to set up a system scheduled task on Windows.
- [ ] Anything you might find.
