# Awkai

Awkai is a small command line utility that uses openAI's chatGPT (gpt-3.5-turbo) to generate awk scripts and executes it.

# Usage

Lets say we have a CSV file `data.csv` containing the following data:
|uid |age|firstname|surname |email |
|----------------|---|---------|-----------|--------------------------------|
|D89S1RGYQ2X5RFZ2|37 |Erma |Trimble |danette.atkins@gmail.com |
|3YJEFUMCHBIELTQ2|11 |Ina |Fitch |esta.callahan69165@carry.com |
|CSGVFK27KZ6NUBR9|91 |Bula |Adkins |laurice98@hotmail.com |
|RZMNB86BSPS4BHZ8|12 |Ginger |Olive |cleveland_dagostino255@birth.com|
|PSIS139HGUIRXG2Y|12 |Mozella |Starr |kerri.bair49453@printing.com |
|20CHPGG763DJZNMF|80 |Dwain |Burt |mitchel5485@hotmail.com |

You can now use awkai to process this data:

```bash
cat data.csv | awkai "Return a new csv with only people using hotmail"
```

This asks chatGPT for an awk script using your query and the first lines of the data as sample data. This script is used and returns the following data:
|uid |age|firstname|surname |email |
|----------------|---|---------|-----------|--------------------------------|
|CSGVFK27KZ6NUBR9|91 |Bula |Adkins |laurice98@hotmail.com |
|20CHPGG763DJZNMF|80 |Dwain |Burt |mitchel5485@hotmail.com |

You can even chain these commands to find the youngest person using hotmail:

```bash
cat data.csv | awkai "Return a new csv with only people using hotmail" | awkai "Find the youngest person"
The youngest person is: Dwain Burt with age 80
```

Note that commands are cached, so it won't contact openAI again for the first command.

# Flags

- `--no-cache`: Sometimes you get unexpected results and you want to try again. Normally commands are cached so you will get the same result every time. If the output is wrong however, you might want to use this flag to create a new script.
- `--dry`: This flag is useful if you want to see the generated awk script before executing it to check if it looks correct.
- `--debug`: Debug is similar to `--dry` but it does execute the awk script. It wil output both the script and the output.
- `--lines`: Lines is used to tune the amount of sample lines that are given to the LLM to use when generating scripts. 10 Lines are used by default.

# Installation

Clone the repository and create the awkai executable using:

```bash
go build *.go
```

Make sure `awkai` is somewhere in your path to use it anywhere.

Another option is use `go install` to put the executable in your `$GOPATH` immediately:

```bash
go build *.go
```

# Notes

- It checks if `gawk` is installed and uses that if it is present. If not, it will check for `awk` and use that if present.
- In the data folder is some test data that you can use for testing.
