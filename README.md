# pushmoi
Send message through Pushbullet

[Pushbullet](https://blog.pushbullet.com/) is a pub/sub service for cross
device data exchange.

## Getting started
You need to authorize this client to perform Pushbullet actions on your behalf.

Execute `pushmoi init` and follow the instructions on the command line
**pushmoi** will setup a web server at **tcp:8080** and await OAuth response.

Upon success you should see your `access_token` on the web UI.  Check back at
the command line to review initialize status.

## Fist step: Review registered devices
Once we had a successful authorization, we can go review our settings and
profile by executing `pushmoi ls`.

Here is an example output:
```
+------------------+---------+-------+--------+
|       NAME       |  TYPE   |  SMS  | ACTIVE |
+------------------+---------+-------+--------+
| Asus Nexus 7     | tablet  | false | true   |
| Motorola Nexus 6 | phone   | true  | true   |
| Chrome           | browser | false | true   |
+------------------+---------+-------+--------+
```

## Second step: set default push target (Recommended)
Setup your default target my executing `pushmoi set default [device name]`,
which you found by `pushmoi ls`

Review your setting by running `pushmoi get default`

## Last step: push a message
There are two ways to push a message:
- Raw text message
- Message formated by template and context

To push a raw text message, execute `pushmoi . [your text message]`.  Note that
the message itself is treated as a single argument, so quote where necessary.

To push a templated message, execute
`pushmoi [template file] [raw text | json encoded string]`

The syntax for our template is documented under
[html/template](https://golang.org/pkg/html/template/)

You could attach **pushmoi** to command line pipeline and have it read output
for your.  Note that there is a limit on the size of the payload encforced by
Pushbullet.

To push through pipeline output, execute
`your-command-or-script | pushmoi [template file] -` to force **pushmoi** to
consume stdin.

## Select a target
There are four ways to select your target:
- default
- email
- device name
- all of your registered devices

If you push message without specification, **pushmoi** pushes to your default
push target.  If a default is not designated, `all` is used.

### All of your registered devices
`pushmoi --all [template] [message]`

### To a device
`pushmoi --device [name] [template] [message]`

### Email
`pushmoi --email [email] [template] [message]`

## Update your settings
If you registered or removed devices, you should update your settings by
executing `pushmoi sync`

If you had revoked **pushmoi** permission, or that the `access_token` was lost,
you should execute `pushmoi init` to restart authorization.
