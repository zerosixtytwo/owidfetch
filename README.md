## owidfetch

> owidfetch is a quick CLI utility that fetches data contained in the JSON
> file that holds pandemic information for all countries contained in the 
> [OWID Covid-19 Repository](https://github.com/owid/covid-19-data)
> and inserts it into a database specified by the user.

### Installation
To install __owidfetch__, open up your terminal and run:
```shell
user@host: ~$ go get -u https://github.com/zerosixtytwo/owidfetch
```
The `owidfetch` binary will be installed in your `$GOPATH/bin` directory, which 
by default is `$USER/go/bin`.

### Usage
Before starting to use __owidfetch__, you will have to create a configuration
file. By default, owidfetch will search for a file called *owidf.conf.yaml* in your
current directory. You can specify a custom configuration file path by using the `-c`
flag. An example configuration file looks like this:
```yaml
# database conn.
DB_DSN: username:password@tcp(127.0.0.1:3306)/dbname
# owid repo.
OWID_DATA_URL: https://raw.githubusercontent.com/owid/covid-19-data/master/public/data/latest/owid-covid-latest.json
```
Once you configured your environment, just open your terminal and run:
```shell
user@host: ~$ $GOPATH/bin/owidfetch # by the way, feel free to move the binary to another location.
```
And all the current data will be stored into the database.  
You could create a cron job in your system that automatically fetches data every x hours.  
An SSD is recommended :)

### Note.
This repository was born as a project for an exam, I had to obtain Covid-19 data
from the internet and store it, then i would have to create a web API that returns
it to an Angular frontend.  
__owidfetch__ will insert ALL the data it fetches from the OWID repository into
your database, thus you should be worried about space since this program is going
to create beast tables with more than 50 columns for every continent. If you want
to omit some of this data you will have to make a fork of owidfetch and edit it
by yourself (good luck since columns and data is inserted using reflection haha).
I decided to separate data by continent/area and not to put everything into a single
table because I think this will affect query execution performance, since if you want
to know data for Italy, for example, you would just look into the `owid_details_europe`
table, which only contains data for European locations.

### Credits for Covid-19 data.
All the credits for fetching pandemic information from official sources and
packing it into easy to parse formats and updating data on a hourly basis
go to the [OWID](https://github.com/owid) group, thank you for making
access to crucial information easier.
_______________________________________________________________________
__Current version:__ 1.25, __Status:__ Stable, Beta.