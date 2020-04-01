[![License](https://img.shields.io/badge/License-Apache2-blue.svg)](https://www.apache.org/licenses/LICENSE-2.0)
[![Version](https://d25lcipzij17d.cloudfront.net/badge.svg?id=go&type=5&v=0.1.0)](https://github.com/wallyqs/covid19mx/releases/tag/v0.1.0)

# covid19mx - Data Tool

Herramienta para obtener los datos actualizados sobre la situación de COVID19 en México.

*Fuente de datos*: http://ncov.sinave.gob.mx/mapa.aspx

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
```

## Demo

[![asciicast](https://asciinema.org/a/hzXbEACTJDSlY9jgzNvBKdQzm.svg)](https://asciinema.org/a/hzXbEACTJDSlY9jgzNvBKdQzm)
