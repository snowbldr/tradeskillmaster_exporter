# tsm_exporter

To use this, first find your TradeSkillMaster AppHelper's AppData.lua file.
On windows, for classic, this is likely at: "C:\Program Files (x86)\World of Warcraft\_classic_\Interface\AddOns\TradeSkillMaster_AppHelper\AppData.lua"

Open this file up, and get the name of the realm you'd like to export data for.

The lines you're looking for start with `select(2, ...).LoadData("AUCTIONDB_MARKET_DATA",`

The realm is the bit immediately after "AUCTIONDB_MARKET_DATA"
For example, my file contains these lines:
```
select(2, ...).LoadData("AUCTIONDB_MARKET_DATA","Sulfuras-Horde",[[return {downloadTime=1664767471,...
select(2, ...).LoadData("AUCTIONDB_MARKET_DATA","BCC-US",[[return {downloadTime=1664752049,...
```

You want to choose either "Sulfuras-Horde" or "BCC-US"

After you have that, make a folder to write the files to on your computer wherever you want.

If you're feeling unimaginative, use `C:\tsm_data`

Now that you have all the input data, install go (https://go.dev/doc/install), checkout this project, then run go install from root of this project.
```go install```

This will install the tsm_exporter program to your GOPATH.

If you've already added your GOPATH to your PATH, you should be able to run tsm_exporter from any terminal.

If you haven't, see here https://www.freecodecamp.org/news/setting-up-go-programming-language-on-windows-f02c8c14e2f/#:~:text=Create%20the%20GOPATH%20variable%20and,System%20Variables%20click%20on%20New.&text=To%20check%20that%20your%20path,%25%E2%80%9D%20on%20the%20command%20line.

or google `add GOPATH to PATH environment mac` or `add GOPATH to PATH environment linux`

Once that's done, run tsm_exporter and pass the 3 arguments you collected before hand.

For example, 
```
tsm_exporter "C:\Program Files (x86)\World of Warcraft\_classic_\Interface\AddOns\TradeSkillMaster_AppHelper\AppData.lua" sulfuras-horde "C:\tsm_data"
```
