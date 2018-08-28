# simgetter.Getter

### Create mode
* Gets using an http.Getter
* Saves page contents and stats to a filesystem database

### Replay mode
* Gets from filesystem database
* Waits a duration based on actual duration (perhaps add some noise in this?)
* Returns as if it had been got from http

