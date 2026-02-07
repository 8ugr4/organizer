# organizer
Lightweight CLI tool written in Go that copies files from a source directory to a destination directory while automatically organizing them into folders by type (audio, images, video, documents, other).

### Requirements

`organized-copy` uses [go-exiftool](https://github.com/barasher/go-exiftool)

**go-exiftool** needs [ExifTool](https://www.sno.phy.queensu.ca/~phil/exiftool/) to be installed.

- On Debian : `sudo apt-get install exiftool`


### Example Usage
Get the binary from releases, or clone the repository and `make build`.

```shell
  # Dry-run: show what will be copied
  ./organizer org-dir --src ~/Downloads --dst ~/Sorted --dry-run

  # Perform copy with log file including status of copy process of every single file and dir
  ./organizer org-dir --src ~/Downloads --dst ~/Sorted --log=~/logfile.csv

  # sha256-validation (optional)
  ./cmpDirs.sh ~/Downloads ~/Sorted
```


## Features
- if multiple files exist with same name + extension, new files get `_number` after first one.
- if user doesn't set a destination path, auto destination path is source path + `_cp` in same directory.
- User can set a rule-set, defining which files will go to which destination.
- Rules have sort option, which puts the files in separate directories depending on their creation date.
- **WIP** 'name_contains' and 'priority_order' is still under development.

## Example

`testDir_cp`: categorized and sorted directory after the run.

```shell
# note:  added ├ between directories for readability purposes.

├── applications
├── archives
├── audios
├── documents
├   ├
│   ├── dummy.txt
│   ├── dummy_1.txt
│   ├── pdf
│   │   ├── asdfgherh4413.pdf
│   │   ├── dfgdsfg.pdf
│   │   ├── dfgdsfg_1.pdf
│   │   ├── feedback.pdf
│   │   ├── feedback2.pdf
│   │   ├── feedback_1.pdf
│   │   ├── feedback_2.pdf
│   │   ├── fsdfgkn.pdf
│   │   └── sofjngongf24214443.pdf
├   ├
│   ├── xyz1.csv
│   └── xyz2.csv
├
├── images
│   ├── 2020
│   │   ├── 10
│   │   │   └── IMG_20201010_170402.jpg
│   │   └── 11
│   │       └── IMG_20201101_174805.jpg
│   ├── 2021
│   │   ├── 03
│   │   │   └── IMG_20210324_200840.jpg
│   │   └── 05
│   │       ├── IMG_20210509_122514.jpg
│   │       ├── IMG_20210511_174630.jpg
│   │       ├── IMG_20210513_173159.jpg
│   │       ├── IMG_20210513_222755.jpg
│   │       ├── IMG_20210517_102755.jpg
│   │       ├── IMG_20210518_001727.jpg
│   │       └── IMG_20210519_203223.jpg
│   ├── dummasdasdy.png
│   ├── dummy.jpg
│   ├── dummy.png
│   ├── dummy2.jpg
│   ├── dummy_1.jpg
│   └── dummy_1.png
├
├── special
├── special2
├── unknown
│   └── fil1
├
└── videos
    └── mp4
        └── abc.mp4

```
