# Golang project - terminal todos app

## Temporary project that will be divided into API and Terminal app

Now to start using app u should start ***main.exe*** from app directory and type `command` to set command `todos` that can be executed anywhere from console that will start the app.

| command    | arguments                     | description                                                              |
|------------|-------------------------------|--------------------------------------------------------------------------|
| exit       |                               |                                                                          |
| help       |                               | prints help                                                              |
| command    |                               | program can be executed from any directory using `todos`                 |
| colors     | 1 / 0 / enable / disable      | using to enable or disable color usage in program                        |
| ls // list |                               | list all stored todos                                                    |
| add        | {Title} {Text} {Deadline} (t) | adds new todo, in case you enter duration {\_}h{\_}m type "t" in the end |
| delete     | {ID}                          | deletes todo                                                             |
| edit       | {ID} {Field} {Value}          | edits todo                                                               |

_datetime format is: dd.MM_hh:mm (d - day, M - month, h - hour, m - minute)_

_duration format is: {\_}h{\_}m (for example 12h30m, or 1h1m, but not 1d12h)_

> Calling command `command` causes program sleeping, you should stop program with `Ctrl + C`, but `todos` is set.

### Coming soon
- Saving data on virtual server.
- Deleting multiple todos using 1 command.
- Clear method (deletes all todos with "done" state)
- Sort method (sorts listed todos using some field)
- Customisation (such as colors, log prefixes, and other)

### Examples
![example](./todo-ex.png)
