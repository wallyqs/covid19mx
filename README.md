[![License](https://img.shields.io/badge/License-Apache2-blue.svg)](https://www.apache.org/licenses/LICENSE-2.0)
[![Version](https://d25lcipzij17d.cloudfront.net/badge.svg?id=go&type=5&v=0.1.4)](https://github.com/wallyqs/covid19mx/releases/tag/v0.1.4)

# covid19mx - Data Tool

Herramienta para obtener los datos actualizados sobre la situación de COVID19 en México 🇲🇽. Para descargar busca el archivo para tu plataforma [aquí](https://github.com/wallyqs/covid19mx/releases).

*Fuente de datos*: https://ncov.sinave.gob.mx/mapa.aspx

```sh
$ covid19mx -h
Usage: covid19mx [options...]

  -h	Show help
  -o string
    	Export format (options: json, csv, table)
  -source string
    	Source
  -version
    	Show version
      
$ covid19mx
|----------------------|-----------------|-----------------|-------------------|---------|
| Estado               | Casos Positivos | Casos Negativos | Casos Sospechosos | Decesos |
|----------------------|-----------------|-----------------|-------------------|---------|
| Aguascalientes       | 36              | 278             | 81                | 0       |
| Baja California      | 35              | 248             | 174               | 0       |
| Baja California Sur  | 17              | 108             | 35                | 0       |
| Campeche             | 5               | 19              | 2                 | 0       |
| Coahuila             | 44              | 242             | 140               | 1       |
| Colima               | 2               | 30              | 8                 | 0       |
| Chiapas              | 13              | 72              | 31                | 0       |
| Chihuahua            | 7               | 52              | 16                | 0       |
| Ciudad de México     | 234             | 695             | 517               | 8       |
| Durango              | 7               | 60              | 18                | 1       |
| Guanajuato           | 46              | 539             | 130               | 0       |
| Guerrero             | 15              | 72              | 80                | 0       |
| Hidalgo              | 19              | 151             | 66                | 3       |
| Jalisco              | 94              | 628             | 377               | 3       |
| México               | 149             | 395             | 371               | 1       |
| Michoacán            | 21              | 125             | 70                | 1       |
| Morelos              | 7               | 72              | 34                | 1       |
| Nayarit              | 6               | 32              | 19                | 0       |
| Nuevo León           | 76              | 559             | 148               | 0       |
| Oaxaca               | 14              | 84              | 50                | 1       |
| Puebla               | 81              | 252             | 121               | 1       |
| Queretaro            | 29              | 171             | 61                | 1       |
| Quintana Roo         | 47              | 157             | 43                | 1       |
| San Luis Potosí      | 25              | 243             | 83                | 2       |
| Sinaloa              | 27              | 150             | 125               | 3       |
| Sonora               | 17              | 118             | 98                | 0       |
| Tabasco              | 48              | 151             | 178               | 0       |
| Tamaulipas           | 8               | 64              | 54                | 0       |
| Tlaxcala             | 4               | 87              | 64                | 0       |
| Veracruz             | 27              | 159             | 266               | 1       |
| Yucatán              | 49              | 151             | 25                | 0       |
| Zacatecas            | 6               | 118             | 26                | 0       |
|----------------------|-----------------|-----------------|-------------------|---------|
| TOTAL                | 1215            | 6282            | 3511              | 29      |
|----------------------|-----------------|-----------------|-------------------|---------|

$ covid19mx --since yesterday
|----------------------|-----------------|-----------------|-------------------|-----------|
| Estado               | Casos Positivos | Casos Negativos | Casos Sospechosos | Decesos   |
|----------------------|-----------------|-----------------|-------------------|-----------|
| Aguascalientes       | 0     (36)      | 43    (321)     | -24   (57)        | 0     (0) |
| Baja California      | 2     (37)      | 34    (282)     | 11    (185)       | 0     (0) |
| Baja California Sur  | 1     (18)      | 3     (111)     | 10    (45)        | 2     (2) |
| Campeche             | 0     (5)       | 0     (19)      | 7     (9)         | 0     (0) |
| Coahuila             | 13    (57)      | 6     (248)     | 22    (162)       | 1     (2) |
| Colima               | 1     (3)       | 1     (31)      | 1     (9)         | 0     (0) |
| Chiapas              | 1     (14)      | 13    (85)      | 5     (36)        | 0     (0) |
| Chihuahua            | 4     (11)      | -1    (51)      | 6     (22)        | 0     (0) |
| Ciudad de México     | 62    (296)     | 92    (787)     | 74    (591)       | 0     (8) |
| Durango              | 0     (7)       | 1     (61)      | 6     (24)        | 0     (1) |
| Guanajuato           | 3     (49)      | 58    (597)     | 43    (173)       | 0     (0) |
| Guerrero             | 2     (17)      | 14    (86)      | 4     (84)        | 1     (1) |
| Hidalgo              | 2     (21)      | 21    (172)     | -6    (60)        | 0     (3) |
| Jalisco              | 5     (99)      | 31    (659)     | 119   (496)       | 0     (3) |
| México               | 8     (157)     | 103   (498)     | -31   (340)       | 0     (1) |
| Michoacán            | 3     (24)      | 23    (148)     | -9    (61)        | 0     (1) |
| Morelos              | 2     (9)       | 16    (88)      | 2     (36)        | 0     (1) |
| Nayarit              | 2     (8)       | 5     (37)      | 1     (20)        | 1     (1) |
| Nuevo León           | 2     (78)      | 88    (647)     | -39   (109)       | 0     (0) |
| Oaxaca               | 8     (22)      | 30    (114)     | -25   (25)        | 0     (1) |
| Puebla               | 16    (97)      | 42    (294)     | -18   (103)       | 0     (1) |
| Queretaro            | 0     (29)      | 8     (179)     | 6     (67)        | 0     (1) |
| Quintana Roo         | 4     (51)      | 4     (161)     | 7     (50)        | 0     (1) |
| San Luis Potosí      | 6     (31)      | 36    (279)     | -9    (74)        | 0     (2) |
| Sinaloa              | 5     (32)      | 19    (169)     | 7     (132)       | 1     (4) |
| Sonora               | 1     (18)      | 9     (127)     | 21    (119)       | 0     (0) |
| Tabasco              | 4     (52)      | 25    (176)     | 55    (233)       | 1     (1) |
| Tamaulipas           | 1     (9)       | 10    (74)      | 7     (61)        | 0     (0) |
| Tlaxcala             | 1     (5)       | 14    (101)     | 2     (66)        | 0     (0) |
| Veracruz             | 1     (28)      | 26    (185)     | 39    (305)       | 0     (1) |
| Yucatán              | 3     (52)      | 15    (166)     | 9     (34)        | 0     (0) |
| Zacatecas            | 0     (6)       | 2     (120)     | 13    (39)        | 1     (1) |
|----------------------|-----------------|-----------------|-------------------|-----------|
| TOTAL                | 163             | 791             | -2455             | 8         |
|----------------------|-----------------|-----------------|-------------------|-----------|
```

## Demo

[![asciicast](https://asciinema.org/a/hzXbEACTJDSlY9jgzNvBKdQzm.svg)](https://asciinema.org/a/hzXbEACTJDSlY9jgzNvBKdQzm)
