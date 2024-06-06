# Changelog

## v0.1

* Initial release
* Added support for multiple scheduler methods:
  * Cron string (with seconds)
  * Interval (duration)
  * At (time)
* Tasks:
  * Built-in http request methods (get,post)
  * Can invoke external commands (same as cron)
    * Supports setting environments variable
  * Can be configured to use any terminal/shell you want such as sh,bash,nu,cmd,powershell,...
  * Can have retries per task (on commands exit-code !=0 or http request errors)
* Support for hooks, for both when tasks fail or they finish successfully (structure is the same as a task)
* Multiple scheduler for each job
* Multiple task for each job
* Logging:
  * File logging
  * ansi/plain/json log formatter
