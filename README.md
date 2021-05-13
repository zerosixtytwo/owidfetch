## owidfetch

> owidfetch is a quick CLI utility that fetches data contained in the JSON
> file that holds pandemic information for all countries contained in the 
> [OWID Covid-19 Repository](https://github.com/owid/covid-19-data)
> and inserts it into a database specified by the user.

### Note.
This repository was born as a project for an exam, I had to obtain Covid-19 data
from the internet and store it, then i would have to create a web API that returns
it to an Angular frontend.  
The stored data is not complete since some values fetched from the OWID repository
have been omitted. The data fetched from the OWID repository would consist in something
like 50+ columns in the database and that, at least to me, seems a bit crazy, so i 
decided to store only most important information.  
If you have any idea on how to store all this data feel free to submit a Pull Request.

### Credits for Covid-19 data.
All the credits for fetching pandemic information from official sources and
packing it into easy to parse formats and updating data on a hourly basis
go to the [OWID](https://github.com/owid) group, thank you for making
access to crucial information easier.
_______________________________________________________________________
__Current version:__ 1.25, __Status:__ Stable, Alpha.