package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	// Used to detect where the actual data is being located, this
	// seems to change day to day.
	sinaveURL = "https://covid19.sinave.gob.mx/mapa.aspx"

	// Latest data will be usually found in one of the following urls.
	sinaveURLA = "https://covid19.sinave.gob.mx/Mapa.aspx/Grafica22"
	sinaveURLB = "https://covid19.sinave.gob.mx/Mapa.aspx/Grafica23"

	// repoURL can be used to fetch previous days date.
	repoURL = "https://wallyqs.github.io/covid19mx/data/"

	// attackRateURL is the url with the info with the
	//
	// See: https://en.wikipedia.org/wiki/Attack_rate
	//      https://es.wikipedia.org/wiki/Incidencia
	//
	attackRateURL = "https://covid19.sinave.gob.mx/Mapatasas.aspx/Grafica22"
)

const (
	version     = "0.4.0"
	releaseDate = "April 22th, 2020"

	// municipalURL is the url from where we can get the data at a
	// municipal level.
	municipalURL = "https://coronavirus.gob.mx/fHDMap/info/getInfoMun.php"
)

var (
	ErrSourceNotFound = errors.New("Could not find datasource!")

	// StatesMap maps the name of a state to an id.
	StatesMap map[string]string = map[string]string{
		"01": "Aguascalientes",
		"02": "Baja California",
		"03": "Baja California Sur",
		"04": "Campeche",
		"05": "Coahuila",
		"06": "Colima",
		"07": "Chiapas",
		"08": "Chihuahua",
		"09": "Ciudad de México",
		"10": "Durango",
		"11": "Guanajuato",
		"12": "Guerrero",
		"13": "Hidalgo",
		"14": "Jalisco",
		"15": "México",
		"16": "Michoacán",
		"17": "Morelos",
		"18": "Nayarit",
		"19": "Nuevo León",
		"20": "Oaxaca",
		"21": "Puebla",
		"22": "Queretaro",
		"23": "Quintana Roo",
		"24": "San Luis Potosí",
		"25": "Sinaloa",
		"26": "Sonora",
		"27": "Tabasco",
		"28": "Tamaulipas",
		"29": "Tlaxcala",
		"30": "Veracruz",
		"31": "Yucatán",
		"32": "Zacatecas",
	}

	// MunicipiosMexico is the
	MunicipiosMexico = map[string]MunicipioDetail{
		"01001": {"01", "Aguascalientes"},
		"01002": {"01", "Asientos"},
		"01003": {"01", "Calvillo"},
		"01004": {"01", "Cosío"},
		"01005": {"01", "Jesús María"},
		"01006": {"01", "Pabellón de Arteaga"},
		"01007": {"01", "Rincón de Romos"},
		"01008": {"01", "San José de Gracia"},
		"01009": {"01", "Tepezalá"},
		"01010": {"01", "El Llano"},
		"01011": {"01", "San Francisco de los Romo"},
		"02001": {"02", "Ensenada"},
		"02002": {"02", "Mexicali"},
		"02003": {"02", "Tecate"},
		"02004": {"02", "Tijuana"},
		"02005": {"02", "Playas de Rosarito"},
		"03001": {"03", "Comondú"},
		"03002": {"03", "Mulegé"},
		"03003": {"03", "La Paz"},
		"03008": {"03", "Los Cabos"},
		"03009": {"03", "Loreto"},
		"04001": {"04", "Calkiní"},
		"04002": {"04", "Campeche"},
		"04003": {"04", "Carmen"},
		"04004": {"04", "Champotón"},
		"04005": {"04", "Hecelchakán"},
		"04006": {"04", "Hopelchén"},
		"04007": {"04", "Palizada"},
		"04008": {"04", "Tenabo"},
		"04009": {"04", "Escárcega"},
		"04010": {"04", "Calakmul"},
		"04011": {"04", "Candelaria"},
		"05001": {"05", "Abasolo"},
		"05002": {"05", "Acuña"},
		"05003": {"05", "Allende"},
		"05004": {"05", "Arteaga"},
		"05005": {"05", "Candela"},
		"05006": {"05", "Castaños"},
		"05007": {"05", "Cuatro Ciénegas"},
		"05008": {"05", "Escobedo"},
		"05009": {"05", "Francisco I. Madero"},
		"05010": {"05", "Frontera"},
		"05011": {"05", "General Cepeda"},
		"05012": {"05", "Guerrero"},
		"05013": {"05", "Hidalgo"},
		"05014": {"05", "Jiménez"},
		"05015": {"05", "Juárez"},
		"05016": {"05", "Lamadrid"},
		"05017": {"05", "Matamoros"},
		"05018": {"05", "Monclova"},
		"05019": {"05", "Morelos"},
		"05020": {"05", "Múzquiz"},
		"05021": {"05", "Nadadores"},
		"05022": {"05", "Nava"},
		"05023": {"05", "Ocampo"},
		"05024": {"05", "Parras"},
		"05025": {"05", "Piedras Negras"},
		"05026": {"05", "Progreso"},
		"05027": {"05", "Ramos Arizpe"},
		"05028": {"05", "Sabinas"},
		"05029": {"05", "Sacramento"},
		"05030": {"05", "Saltillo"},
		"05031": {"05", "San Buenaventura"},
		"05032": {"05", "San Juan de Sabinas"},
		"05033": {"05", "San Pedro"},
		"05034": {"05", "Sierra Mojada"},
		"05035": {"05", "Torreón"},
		"05036": {"05", "Viesca"},
		"05037": {"05", "Villa Unión"},
		"05038": {"05", "Zaragoza"},
		"06001": {"06", "Armería"},
		"06002": {"06", "Colima"},
		"06003": {"06", "Comala"},
		"06004": {"06", "Coquimatlán"},
		"06005": {"06", "Cuauhtémoc"},
		"06006": {"06", "Ixtlahuacán"},
		"06007": {"06", "Manzanillo"},
		"06008": {"06", "Minatitlán"},
		"06009": {"06", "Tecomán"},
		"06010": {"06", "Villa de Álvarez"},
		"07001": {"07", "Acacoyagua"},
		"07002": {"07", "Acala"},
		"07003": {"07", "Acapetahua"},
		"07004": {"07", "Altamirano"},
		"07005": {"07", "Amatán"},
		"07006": {"07", "Amatenango de la Frontera"},
		"07007": {"07", "Amatenango del Valle"},
		"07008": {"07", "Angel Albino Corzo"},
		"07009": {"07", "Arriaga"},
		"07010": {"07", "Bejucal de Ocampo"},
		"07011": {"07", "Bella Vista"},
		"07012": {"07", "Berriozábal"},
		"07013": {"07", "Bochil"},
		"07014": {"07", "El Bosque"},
		"07015": {"07", "Cacahoatán"},
		"07016": {"07", "Catazajá"},
		"07017": {"07", "Cintalapa"},
		"07018": {"07", "Coapilla"},
		"07019": {"07", "Comitán de Domínguez"},
		"07020": {"07", "La Concordia"},
		"07021": {"07", "Copainalá"},
		"07022": {"07", "Chalchihuitán"},
		"07023": {"07", "Chamula"},
		"07024": {"07", "Chanal"},
		"07025": {"07", "Chapultenango"},
		"07026": {"07", "Chenalhó"},
		"07027": {"07", "Chiapa de Corzo"},
		"07028": {"07", "Chiapilla"},
		"07029": {"07", "Chicoasén"},
		"07030": {"07", "Chicomuselo"},
		"07031": {"07", "Chilón"},
		"07032": {"07", "Escuintla"},
		"07033": {"07", "Francisco León"},
		"07034": {"07", "Frontera Comalapa"},
		"07035": {"07", "Frontera Hidalgo"},
		"07036": {"07", "La Grandeza"},
		"07037": {"07", "Huehuetán"},
		"07038": {"07", "Huixtán"},
		"07039": {"07", "Huitiupán"},
		"07040": {"07", "Huixtla"},
		"07041": {"07", "La Independencia"},
		"07042": {"07", "Ixhuatán"},
		"07043": {"07", "Ixtacomitán"},
		"07044": {"07", "Ixtapa"},
		"07045": {"07", "Ixtapangajoya"},
		"07046": {"07", "Jiquipilas"},
		"07047": {"07", "Jitotol"},
		"07048": {"07", "Juárez"},
		"07049": {"07", "Larráinzar"},
		"07050": {"07", "La Libertad"},
		"07051": {"07", "Mapastepec"},
		"07052": {"07", "Las Margaritas"},
		"07053": {"07", "Mazapa de Madero"},
		"07054": {"07", "Mazatán"},
		"07055": {"07", "Metapa"},
		"07056": {"07", "Mitontic"},
		"07057": {"07", "Motozintla"},
		"07058": {"07", "Nicolás Ruíz"},
		"07059": {"07", "Ocosingo"},
		"07060": {"07", "Ocotepec"},
		"07061": {"07", "Ocozocoautla de Espinosa"},
		"07062": {"07", "Ostuacán"},
		"07063": {"07", "Osumacinta"},
		"07064": {"07", "Oxchuc"},
		"07065": {"07", "Palenque"},
		"07066": {"07", "Pantelhó"},
		"07067": {"07", "Pantepec"},
		"07068": {"07", "Pichucalco"},
		"07069": {"07", "Pijijiapan"},
		"07070": {"07", "El Porvenir"},
		"07071": {"07", "Villa Comaltitlán"},
		"07072": {"07", "Pueblo Nuevo Solistahuacán"},
		"07073": {"07", "Rayón"},
		"07074": {"07", "Reforma"},
		"07075": {"07", "Las Rosas"},
		"07076": {"07", "Sabanilla"},
		"07077": {"07", "Salto de Agua"},
		"07078": {"07", "San Cristóbal de las Casas"},
		"07079": {"07", "San Fernando"},
		"07080": {"07", "Siltepec"},
		"07081": {"07", "Simojovel"},
		"07082": {"07", "Sitalá"},
		"07083": {"07", "Socoltenango"},
		"07084": {"07", "Solosuchiapa"},
		"07085": {"07", "Soyaló"},
		"07086": {"07", "Suchiapa"},
		"07087": {"07", "Suchiate"},
		"07088": {"07", "Sunuapa"},
		"07089": {"07", "Tapachula"},
		"07090": {"07", "Tapalapa"},
		"07091": {"07", "Tapilula"},
		"07092": {"07", "Tecpatán"},
		"07093": {"07", "Tenejapa"},
		"07094": {"07", "Teopisca"},
		"07096": {"07", "Tila"},
		"07097": {"07", "Tonalá"},
		"07098": {"07", "Totolapa"},
		"07099": {"07", "La Trinitaria"},
		"07100": {"07", "Tumbalá"},
		"07101": {"07", "Tuxtla Gutiérrez"},
		"07102": {"07", "Tuxtla Chico"},
		"07103": {"07", "Tuzantán"},
		"07104": {"07", "Tzimol"},
		"07105": {"07", "Unión Juárez"},
		"07106": {"07", "Venustiano Carranza"},
		"07107": {"07", "Villa Corzo"},
		"07108": {"07", "Villaflores"},
		"07109": {"07", "Yajalón"},
		"07110": {"07", "San Lucas"},
		"07111": {"07", "Zinacantán"},
		"07112": {"07", "San Juan Cancuc"},
		"07113": {"07", "Aldama"},
		"07114": {"07", "Benemérito de las Américas"},
		"07115": {"07", "Maravilla Tenejapa"},
		"07116": {"07", "Marqués de Comillas"},
		"07117": {"07", "Montecristo de Guerrero"},
		"07118": {"07", "San Andrés Duraznal"},
		"07119": {"07", "Santiago el Pinar"},
		"08001": {"08", "Ahumada"},
		"08002": {"08", "Aldama"},
		"08003": {"08", "Allende"},
		"08004": {"08", "Aquiles Serdán"},
		"08005": {"08", "Ascensión"},
		"08006": {"08", "Bachíniva"},
		"08007": {"08", "Balleza"},
		"08008": {"08", "Batopilas"},
		"08009": {"08", "Bocoyna"},
		"08010": {"08", "Buenaventura"},
		"08011": {"08", "Camargo"},
		"08012": {"08", "Carichí"},
		"08013": {"08", "Casas Grandes"},
		"08014": {"08", "Coronado"},
		"08015": {"08", "Coyame del Sotol"},
		"08016": {"08", "La Cruz"},
		"08017": {"08", "Cuauhtémoc"},
		"08018": {"08", "Cusihuiriachi"},
		"08019": {"08", "Chihuahua"},
		"08020": {"08", "Chínipas"},
		"08021": {"08", "Delicias"},
		"08022": {"08", "Dr. Belisario Domínguez"},
		"08023": {"08", "Galeana"},
		"08024": {"08", "Santa Isabel"},
		"08025": {"08", "Gómez Farías"},
		"08026": {"08", "Gran Morelos"},
		"08027": {"08", "Guachochi"},
		"08028": {"08", "Guadalupe"},
		"08029": {"08", "Guadalupe y Calvo"},
		"08030": {"08", "Guazapares"},
		"08031": {"08", "Guerrero"},
		"08032": {"08", "Hidalgo del Parral"},
		"08033": {"08", "Huejotitán"},
		"08034": {"08", "Ignacio Zaragoza"},
		"08035": {"08", "Janos"},
		"08036": {"08", "Jiménez"},
		"08037": {"08", "Juárez"},
		"08038": {"08", "Julimes"},
		"08039": {"08", "López"},
		"08040": {"08", "Madera"},
		"08041": {"08", "Maguarichi"},
		"08042": {"08", "Manuel Benavides"},
		"08043": {"08", "Matachí"},
		"08044": {"08", "Matamoros"},
		"08045": {"08", "Meoqui"},
		"08046": {"08", "Morelos"},
		"08047": {"08", "Moris"},
		"08048": {"08", "Namiquipa"},
		"08049": {"08", "Nonoava"},
		"08050": {"08", "Nuevo Casas Grandes"},
		"08051": {"08", "Ocampo"},
		"08052": {"08", "Ojinaga"},
		"08053": {"08", "Praxedis G. Guerrero"},
		"08054": {"08", "Riva Palacio"},
		"08055": {"08", "Rosales"},
		"08056": {"08", "Rosario"},
		"08057": {"08", "San Francisco de Borja"},
		"08058": {"08", "San Francisco de Conchos"},
		"08059": {"08", "San Francisco del Oro"},
		"08060": {"08", "Santa Bárbara"},
		"08061": {"08", "Satevó"},
		"08062": {"08", "Saucillo"},
		"08063": {"08", "Temósachic"},
		"08064": {"08", "El Tule"},
		"08065": {"08", "Urique"},
		"08066": {"08", "Uruachi"},
		"08067": {"08", "Valle de Zaragoza"},
		"09002": {"09", "Azcapotzalco"},
		"09003": {"09", "Coyoacán"},
		"09004": {"09", "Cuajimalpa de Morelos"},
		"09005": {"09", "Gustavo A. Madero"},
		"09006": {"09", "Iztacalco"},
		"09007": {"09", "Iztapalapa"},
		"09008": {"09", "La Magdalena Contreras"},
		"09009": {"09", "Milpa Alta"},
		"09010": {"09", "Álvaro Obregón"},
		"09011": {"09", "Tláhuac"},
		"09012": {"09", "Tlalpan"},
		"09013": {"09", "Xochimilco"},
		"09014": {"09", "Benito Juárez"},
		"09015": {"09", "Cuauhtémoc"},
		"09016": {"09", "Miguel Hidalgo"},
		"09017": {"09", "Venustiano Carranza"},
		"10001": {"10", "Canatlán"},
		"10002": {"10", "Canelas"},
		"10003": {"10", "Coneto de Comonfort"},
		"10004": {"10", "Cuencamé"},
		"10005": {"10", "Durango"},
		"10006": {"10", "General Simón Bolívar"},
		"10007": {"10", "Gómez Palacio"},
		"10008": {"10", "Guadalupe Victoria"},
		"10009": {"10", "Guanaceví"},
		"10010": {"10", "Hidalgo"},
		"10011": {"10", "Indé"},
		"10012": {"10", "Lerdo"},
		"10013": {"10", "Mapimí"},
		"10014": {"10", "Mezquital"},
		"10015": {"10", "Nazas"},
		"10016": {"10", "Nombre de Dios"},
		"10017": {"10", "Ocampo"},
		"10018": {"10", "El Oro"},
		"10019": {"10", "Otáez"},
		"10020": {"10", "Pánuco de Coronado"},
		"10021": {"10", "Peñón Blanco"},
		"10022": {"10", "Poanas"},
		"10023": {"10", "Pueblo Nuevo"},
		"10024": {"10", "Rodeo"},
		"10025": {"10", "San Bernardo"},
		"10026": {"10", "San Dimas"},
		"10027": {"10", "San Juan de Guadalupe"},
		"10028": {"10", "San Juan del Río"},
		"10029": {"10", "San Luis del Cordero"},
		"10030": {"10", "San Pedro del Gallo"},
		"10031": {"10", "Santa Clara"},
		"10032": {"10", "Santiago Papasquiaro"},
		"10033": {"10", "Súchil"},
		"10034": {"10", "Tamazula"},
		"10035": {"10", "Tepehuanes"},
		"10036": {"10", "Tlahualilo"},
		"10037": {"10", "Topia"},
		"10038": {"10", "Vicente Guerrero"},
		"10039": {"10", "Nuevo Ideal"},
		"11001": {"11", "Abasolo"},
		"11002": {"11", "Acámbaro"},
		"11003": {"11", "San Miguel de Allende"},
		"11004": {"11", "Apaseo el Alto"},
		"11005": {"11", "Apaseo el Grande"},
		"11006": {"11", "Atarjea"},
		"11007": {"11", "Celaya"},
		"11008": {"11", "Manuel Doblado"},
		"11009": {"11", "Comonfort"},
		"11010": {"11", "Coroneo"},
		"11011": {"11", "Cortazar"},
		"11012": {"11", "Cuerámaro"},
		"11013": {"11", "Doctor Mora"},
		"11014": {"11", "Dolores Hidalgo Cuna de la Independencia Nacional"},
		"11015": {"11", "Guanajuato"},
		"11016": {"11", "Huanímaro"},
		"11017": {"11", "Irapuato"},
		"11018": {"11", "Jaral del Progreso"},
		"11019": {"11", "Jerécuaro"},
		"11020": {"11", "León"},
		"11021": {"11", "Moroleón"},
		"11022": {"11", "Ocampo"},
		"11023": {"11", "Pénjamo"},
		"11024": {"11", "Pueblo Nuevo"},
		"11025": {"11", "Purísima del Rincón"},
		"11026": {"11", "Romita"},
		"11027": {"11", "Salamanca"},
		"11028": {"11", "Salvatierra"},
		"11029": {"11", "San Diego de la Unión"},
		"11030": {"11", "San Felipe"},
		"11031": {"11", "San Francisco del Rincón"},
		"11032": {"11", "San José Iturbide"},
		"11033": {"11", "San Luis de la Paz"},
		"11034": {"11", "Santa Catarina"},
		"11035": {"11", "Santa Cruz de Juventino Rosas"},
		"11036": {"11", "Santiago Maravatío"},
		"11037": {"11", "Silao"},
		"11038": {"11", "Tarandacuao"},
		"11039": {"11", "Tarimoro"},
		"11040": {"11", "Tierra Blanca"},
		"11041": {"11", "Uriangato"},
		"11042": {"11", "Valle de Santiago"},
		"11043": {"11", "Victoria"},
		"11044": {"11", "Villagrán"},
		"11045": {"11", "Xichú"},
		"11046": {"11", "Yuriria"},
		"12001": {"12", "Acapulco de Juárez"},
		"12002": {"12", "Ahuacuotzingo"},
		"12003": {"12", "Ajuchitlán del Progreso"},
		"12004": {"12", "Alcozauca de Guerrero"},
		"12005": {"12", "Alpoyeca"},
		"12006": {"12", "Apaxtla"},
		"12007": {"12", "Arcelia"},
		"12008": {"12", "Atenango del Río"},
		"12009": {"12", "Atlamajalcingo del Monte"},
		"12010": {"12", "Atlixtac"},
		"12011": {"12", "Atoyac de Álvarez"},
		"12012": {"12", "Ayutla de los Libres"},
		"12013": {"12", "Azoyú"},
		"12014": {"12", "Benito Juárez"},
		"12015": {"12", "Buenavista de Cuéllar"},
		"12016": {"12", "Coahuayutla de José María Izazaga"},
		"12017": {"12", "Cocula"},
		"12018": {"12", "Copala"},
		"12019": {"12", "Copalillo"},
		"12020": {"12", "Copanatoyac"},
		"12021": {"12", "Coyuca de Benítez"},
		"12022": {"12", "Coyuca de Catalán"},
		"12023": {"12", "Cuajinicuilapa"},
		"12024": {"12", "Cualác"},
		"12025": {"12", "Cuautepec"},
		"12026": {"12", "Cuetzala del Progreso"},
		"12027": {"12", "Cutzamala de Pinzón"},
		"12028": {"12", "Chilapa de Álvarez"},
		"12029": {"12", "Chilpancingo de los Bravo"},
		"12030": {"12", "Florencio Villarreal"},
		"12031": {"12", "General Canuto A. Neri"},
		"12032": {"12", "General Heliodoro Castillo"},
		"12033": {"12", "Huamuxtitlán"},
		"12034": {"12", "Huitzuco de los Figueroa"},
		"12035": {"12", "Iguala de la Independencia"},
		"12036": {"12", "Igualapa"},
		"12037": {"12", "Ixcateopan de Cuauhtémoc"},
		"12038": {"12", "Zihuatanejo de Azueta"},
		"12039": {"12", "Juan R. Escudero"},
		"12040": {"12", "Leonardo Bravo"},
		"12041": {"12", "Malinaltepec"},
		"12042": {"12", "Mártir de Cuilapan"},
		"12043": {"12", "Metlatónoc"},
		"12044": {"12", "Mochitlán"},
		"12045": {"12", "Olinalá"},
		"12046": {"12", "Ometepec"},
		"12047": {"12", "Pedro Ascencio Alquisiras"},
		"12048": {"12", "Petatlán"},
		"12049": {"12", "Pilcaya"},
		"12050": {"12", "Pungarabato"},
		"12051": {"12", "Quechultenango"},
		"12052": {"12", "San Luis Acatlán"},
		"12053": {"12", "San Marcos"},
		"12054": {"12", "San Miguel Totolapan"},
		"12055": {"12", "Taxco de Alarcón"},
		"12056": {"12", "Tecoanapa"},
		"12057": {"12", "Técpan de Galeana"},
		"12058": {"12", "Teloloapan"},
		"12059": {"12", "Tepecoacuilco de Trujano"},
		"12060": {"12", "Tetipac"},
		"12061": {"12", "Tixtla de Guerrero"},
		"12062": {"12", "Tlacoachistlahuaca"},
		"12063": {"12", "Tlacoapa"},
		"12064": {"12", "Tlalchapa"},
		"12065": {"12", "Tlalixtaquilla de Maldonado"},
		"12066": {"12", "Tlapa de Comonfort"},
		"12067": {"12", "Tlapehuala"},
		"12068": {"12", "La Unión de Isidoro Montes de Oca"},
		"12069": {"12", "Xalpatláhuac"},
		"12070": {"12", "Xochihuehuetlán"},
		"12071": {"12", "Xochistlahuaca"},
		"12072": {"12", "Zapotitlán Tablas"},
		"12073": {"12", "Zirándaro"},
		"12074": {"12", "Zitlala"},
		"12075": {"12", "Eduardo Neri"},
		"12076": {"12", "Acatepec"},
		"12077": {"12", "Marquelia"},
		"12078": {"12", "Cochoapa el Grande"},
		"12079": {"12", "José Joaquin de Herrera"},
		"12080": {"12", "Juchitán"},
		"12081": {"12", "Iliatenco"},
		"13001": {"13", "Acatlán"},
		"13002": {"13", "Acaxochitlán"},
		"13003": {"13", "Actopan"},
		"13004": {"13", "Agua Blanca de Iturbide"},
		"13005": {"13", "Ajacuba"},
		"13006": {"13", "Alfajayucan"},
		"13007": {"13", "Almoloya"},
		"13008": {"13", "Apan"},
		"13009": {"13", "El Arenal"},
		"13010": {"13", "Atitalaquia"},
		"13011": {"13", "Atlapexco"},
		"13012": {"13", "Atotonilco el Grande"},
		"13013": {"13", "Atotonilco de Tula"},
		"13014": {"13", "Calnali"},
		"13015": {"13", "Cardonal"},
		"13016": {"13", "Cuautepec de Hinojosa"},
		"13017": {"13", "Chapantongo"},
		"13018": {"13", "Chapulhuacán"},
		"13019": {"13", "Chilcuautla"},
		"13020": {"13", "Eloxochitlán"},
		"13021": {"13", "Emiliano Zapata"},
		"13022": {"13", "Epazoyucan"},
		"13023": {"13", "Francisco I. Madero"},
		"13024": {"13", "Huasca de Ocampo"},
		"13025": {"13", "Huautla"},
		"13026": {"13", "Huazalingo"},
		"13027": {"13", "Huehuetla"},
		"13028": {"13", "Huejutla de Reyes"},
		"13029": {"13", "Huichapan"},
		"13030": {"13", "Ixmiquilpan"},
		"13031": {"13", "Jacala de Ledezma"},
		"13032": {"13", "Jaltocán"},
		"13033": {"13", "Juárez Hidalgo"},
		"13034": {"13", "Lolotla"},
		"13035": {"13", "Metepec"},
		"13036": {"13", "San Agustín Metzquititlán"},
		"13037": {"13", "Metztitlán"},
		"13038": {"13", "Mineral del Chico"},
		"13039": {"13", "Mineral del Monte"},
		"13040": {"13", "La Misión"},
		"13041": {"13", "Mixquiahuala de Juárez"},
		"13042": {"13", "Molango de Escamilla"},
		"13043": {"13", "Nicolás Flores"},
		"13044": {"13", "Nopala de Villagrán"},
		"13045": {"13", "Omitlán de Juárez"},
		"13046": {"13", "San Felipe Orizatlán"},
		"13047": {"13", "Pacula"},
		"13048": {"13", "Pachuca de Soto"},
		"13049": {"13", "Pisaflores"},
		"13050": {"13", "Progreso de Obregón"},
		"13051": {"13", "Mineral de la Reforma"},
		"13052": {"13", "San Agustín Tlaxiaca"},
		"13053": {"13", "San Bartolo Tutotepec"},
		"13054": {"13", "San Salvador"},
		"13055": {"13", "Santiago de Anaya"},
		"13056": {"13", "Santiago Tulantepec de Lugo Guerrero"},
		"13057": {"13", "Singuilucan"},
		"13058": {"13", "Tasquillo"},
		"13059": {"13", "Tecozautla"},
		"13060": {"13", "Tenango de Doria"},
		"13061": {"13", "Tepeapulco"},
		"13062": {"13", "Tepehuacán de Guerrero"},
		"13063": {"13", "Tepeji del Río de Ocampo"},
		"13064": {"13", "Tepetitlán"},
		"13065": {"13", "Tetepango"},
		"13066": {"13", "Villa de Tezontepec"},
		"13067": {"13", "Tezontepec de Aldama"},
		"13068": {"13", "Tianguistengo"},
		"13069": {"13", "Tizayuca"},
		"13070": {"13", "Tlahuelilpan"},
		"13071": {"13", "Tlahuiltepa"},
		"13072": {"13", "Tlanalapa"},
		"13073": {"13", "Tlanchinol"},
		"13074": {"13", "Tlaxcoapan"},
		"13075": {"13", "Tolcayuca"},
		"13076": {"13", "Tula de Allende"},
		"13077": {"13", "Tulancingo de Bravo"},
		"13078": {"13", "Xochiatipan"},
		"13079": {"13", "Xochicoatlán"},
		"13080": {"13", "Yahualica"},
		"13081": {"13", "Zacualtipán de Ángeles"},
		"13082": {"13", "Zapotlán de Juárez"},
		"13083": {"13", "Zempoala"},
		"13084": {"13", "Zimapán"},
		"14001": {"14", "Acatic"},
		"14002": {"14", "Acatlán de Juárez"},
		"14003": {"14", "Ahualulco de Mercado"},
		"14004": {"14", "Amacueca"},
		"14005": {"14", "Amatitán"},
		"14006": {"14", "Ameca"},
		"14007": {"14", "San Juanito de Escobedo"},
		"14008": {"14", "Arandas"},
		"14009": {"14", "El Arenal"},
		"14010": {"14", "Atemajac de Brizuela"},
		"14011": {"14", "Atengo"},
		"14012": {"14", "Atenguillo"},
		"14013": {"14", "Atotonilco el Alto"},
		"14014": {"14", "Atoyac"},
		"14015": {"14", "Autlán de Navarro"},
		"14016": {"14", "Ayotlán"},
		"14017": {"14", "Ayutla"},
		"14018": {"14", "La Barca"},
		"14019": {"14", "Bolaños"},
		"14020": {"14", "Cabo Corrientes"},
		"14021": {"14", "Casimiro Castillo"},
		"14022": {"14", "Cihuatlán"},
		"14023": {"14", "Zapotlán el Grande"},
		"14024": {"14", "Cocula"},
		"14025": {"14", "Colotlán"},
		"14026": {"14", "Concepción de Buenos Aires"},
		"14027": {"14", "Cuautitlán de García Barragán"},
		"14028": {"14", "Cuautla"},
		"14029": {"14", "Cuquío"},
		"14030": {"14", "Chapala"},
		"14031": {"14", "Chimaltitán"},
		"14032": {"14", "Chiquilistlán"},
		"14033": {"14", "Degollado"},
		"14034": {"14", "Ejutla"},
		"14035": {"14", "Encarnación de Díaz"},
		"14036": {"14", "Etzatlán"},
		"14037": {"14", "El Grullo"},
		"14038": {"14", "Guachinango"},
		"14039": {"14", "Guadalajara"},
		"14040": {"14", "Hostotipaquillo"},
		"14041": {"14", "Huejúcar"},
		"14042": {"14", "Huejuquilla el Alto"},
		"14043": {"14", "La Huerta"},
		"14044": {"14", "Ixtlahuacán de los Membrillos"},
		"14045": {"14", "Ixtlahuacán del Río"},
		"14046": {"14", "Jalostotitlán"},
		"14047": {"14", "Jamay"},
		"14048": {"14", "Jesús María"},
		"14049": {"14", "Jilotlán de los Dolores"},
		"14050": {"14", "Jocotepec"},
		"14051": {"14", "Juanacatlán"},
		"14052": {"14", "Juchitlán"},
		"14053": {"14", "Lagos de Moreno"},
		"14054": {"14", "El Limón"},
		"14055": {"14", "Magdalena"},
		"14056": {"14", "Santa María del Oro"},
		"14057": {"14", "La Manzanilla de la Paz"},
		"14058": {"14", "Mascota"},
		"14059": {"14", "Mazamitla"},
		"14060": {"14", "Mexticacán"},
		"14061": {"14", "Mezquitic"},
		"14062": {"14", "Mixtlán"},
		"14063": {"14", "Ocotlán"},
		"14064": {"14", "Ojuelos de Jalisco"},
		"14065": {"14", "Pihuamo"},
		"14066": {"14", "Poncitlán"},
		"14067": {"14", "Puerto Vallarta"},
		"14068": {"14", "Villa Purificación"},
		"14069": {"14", "Quitupan"},
		"14070": {"14", "El Salto"},
		"14071": {"14", "San Cristóbal de la Barranca"},
		"14072": {"14", "San Diego de Alejandría"},
		"14073": {"14", "San Juan de los Lagos"},
		"14074": {"14", "San Julián"},
		"14075": {"14", "San Marcos"},
		"14076": {"14", "San Martín de Bolaños"},
		"14077": {"14", "San Martín Hidalgo"},
		"14078": {"14", "San Miguel el Alto"},
		"14079": {"14", "Gómez Farías"},
		"14080": {"14", "San Sebastián del Oeste"},
		"14081": {"14", "Santa María de los Ángeles"},
		"14082": {"14", "Sayula"},
		"14083": {"14", "Tala"},
		"14084": {"14", "Talpa de Allende"},
		"14085": {"14", "Tamazula de Gordiano"},
		"14086": {"14", "Tapalpa"},
		"14087": {"14", "Tecalitlán"},
		"14088": {"14", "Tecolotlán"},
		"14089": {"14", "Techaluta de Montenegro"},
		"14090": {"14", "Tenamaxtlán"},
		"14091": {"14", "Teocaltiche"},
		"14092": {"14", "Teocuitatlán de Corona"},
		"14093": {"14", "Tepatitlán de Morelos"},
		"14094": {"14", "Tequila"},
		"14095": {"14", "Teuchitlán"},
		"14096": {"14", "Tizapán el Alto"},
		"14097": {"14", "Tlajomulco de Zúñiga"},
		"14098": {"14", "Tlaquepaque"},
		"14099": {"14", "Tolimán"},
		"14100": {"14", "Tomatlán"},
		"14101": {"14", "Tonalá"},
		"14102": {"14", "Tonaya"},
		"14103": {"14", "Tonila"},
		"14104": {"14", "Totatiche"},
		"14105": {"14", "Tototlán"},
		"14106": {"14", "Tuxcacuesco"},
		"14107": {"14", "Tuxcueca"},
		"14108": {"14", "Tuxpan"},
		"14109": {"14", "Unión de San Antonio"},
		"14110": {"14", "Unión de Tula"},
		"14111": {"14", "Valle de Guadalupe"},
		"14112": {"14", "Valle de Juárez"},
		"14113": {"14", "San Gabriel"},
		"14114": {"14", "Villa Corona"},
		"14115": {"14", "Villa Guerrero"},
		"14116": {"14", "Villa Hidalgo"},
		"14117": {"14", "Cañadas de Obregón"},
		"14118": {"14", "Yahualica de González Gallo"},
		"14119": {"14", "Zacoalco de Torres"},
		"14120": {"14", "Zapopan"},
		"14121": {"14", "Zapotiltic"},
		"14122": {"14", "Zapotitlán de Vadillo"},
		"14123": {"14", "Zapotlán del Rey"},
		"14124": {"14", "Zapotlanejo"},
		"14125": {"14", "San Ignacio Cerro Gordo"},
		"15001": {"15", "Acambay"},
		"15002": {"15", "Acolman"},
		"15003": {"15", "Aculco"},
		"15004": {"15", "Almoloya de Alquisiras"},
		"15005": {"15", "Almoloya de Juárez"},
		"15006": {"15", "Almoloya del Río"},
		"15007": {"15", "Amanalco"},
		"15008": {"15", "Amatepec"},
		"15009": {"15", "Amecameca"},
		"15010": {"15", "Apaxco"},
		"15011": {"15", "Atenco"},
		"15012": {"15", "Atizapán"},
		"15013": {"15", "Atizapán de Zaragoza"},
		"15014": {"15", "Atlacomulco"},
		"15015": {"15", "Atlautla"},
		"15016": {"15", "Axapusco"},
		"15017": {"15", "Ayapango"},
		"15018": {"15", "Calimaya"},
		"15019": {"15", "Capulhuac"},
		"15020": {"15", "Coacalco de Berriozábal"},
		"15021": {"15", "Coatepec Harinas"},
		"15022": {"15", "Cocotitlán"},
		"15023": {"15", "Coyotepec"},
		"15024": {"15", "Cuautitlán"},
		"15025": {"15", "Chalco"},
		"15026": {"15", "Chapa de Mota"},
		"15027": {"15", "Chapultepec"},
		"15028": {"15", "Chiautla"},
		"15029": {"15", "Chicoloapan"},
		"15030": {"15", "Chiconcuac"},
		"15031": {"15", "Chimalhuacán"},
		"15032": {"15", "Donato Guerra"},
		"15033": {"15", "Ecatepec de Morelos"},
		"15034": {"15", "Ecatzingo"},
		"15035": {"15", "Huehuetoca"},
		"15036": {"15", "Hueypoxtla"},
		"15037": {"15", "Huixquilucan"},
		"15038": {"15", "Isidro Fabela"},
		"15039": {"15", "Ixtapaluca"},
		"15040": {"15", "Ixtapan de la Sal"},
		"15041": {"15", "Ixtapan del Oro"},
		"15042": {"15", "Ixtlahuaca"},
		"15043": {"15", "Xalatlaco"},
		"15044": {"15", "Jaltenco"},
		"15045": {"15", "Jilotepec"},
		"15046": {"15", "Jilotzingo"},
		"15047": {"15", "Jiquipilco"},
		"15048": {"15", "Jocotitlán"},
		"15049": {"15", "Joquicingo"},
		"15050": {"15", "Juchitepec"},
		"15051": {"15", "Lerma"},
		"15052": {"15", "Malinalco"},
		"15053": {"15", "Melchor Ocampo"},
		"15054": {"15", "Metepec"},
		"15055": {"15", "Mexicaltzingo"},
		"15056": {"15", "Morelos"},
		"15057": {"15", "Naucalpan de Juárez"},
		"15058": {"15", "Nezahualcóyotl"},
		"15059": {"15", "Nextlalpan"},
		"15060": {"15", "Nicolás Romero"},
		"15061": {"15", "Nopaltepec"},
		"15062": {"15", "Ocoyoacac"},
		"15063": {"15", "Ocuilan"},
		"15064": {"15", "El Oro"},
		"15065": {"15", "Otumba"},
		"15066": {"15", "Otzoloapan"},
		"15067": {"15", "Otzolotepec"},
		"15068": {"15", "Ozumba"},
		"15069": {"15", "Papalotla"},
		"15070": {"15", "La Paz"},
		"15071": {"15", "Polotitlán"},
		"15072": {"15", "Rayón"},
		"15073": {"15", "San Antonio la Isla"},
		"15074": {"15", "San Felipe del Progreso"},
		"15075": {"15", "San Martín de las Pirámides"},
		"15076": {"15", "San Mateo Atenco"},
		"15077": {"15", "San Simón de Guerrero"},
		"15078": {"15", "Santo Tomás"},
		"15079": {"15", "Soyaniquilpan de Juárez"},
		"15080": {"15", "Sultepec"},
		"15081": {"15", "Tecámac"},
		"15082": {"15", "Tejupilco"},
		"15083": {"15", "Temamatla"},
		"15084": {"15", "Temascalapa"},
		"15085": {"15", "Temascalcingo"},
		"15086": {"15", "Temascaltepec"},
		"15087": {"15", "Temoaya"},
		"15088": {"15", "Tenancingo"},
		"15089": {"15", "Tenango del Aire"},
		"15090": {"15", "Tenango del Valle"},
		"15091": {"15", "Teoloyucán"},
		"15092": {"15", "Teotihuacán"},
		"15093": {"15", "Tepetlaoxtoc"},
		"15094": {"15", "Tepetlixpa"},
		"15095": {"15", "Tepotzotlán"},
		"15096": {"15", "Tequixquiac"},
		"15097": {"15", "Texcaltitlán"},
		"15098": {"15", "Texcalyacac"},
		"15099": {"15", "Texcoco"},
		"15100": {"15", "Tezoyuca"},
		"15101": {"15", "Tianguistenco"},
		"15102": {"15", "Timilpan"},
		"15103": {"15", "Tlalmanalco"},
		"15104": {"15", "Tlalnepantla de Baz"},
		"15105": {"15", "Tlatlaya"},
		"15106": {"15", "Toluca"},
		"15107": {"15", "Tonatico"},
		"15108": {"15", "Tultepec"},
		"15109": {"15", "Tultitlán"},
		"15110": {"15", "Valle de Bravo"},
		"15111": {"15", "Villa de Allende"},
		"15112": {"15", "Villa del Carbón"},
		"15113": {"15", "Villa Guerrero"},
		"15114": {"15", "Villa Victoria"},
		"15115": {"15", "Xonacatlán"},
		"15116": {"15", "Zacazonapan"},
		"15117": {"15", "Zacualpan"},
		"15118": {"15", "Zinacantepec"},
		"15119": {"15", "Zumpahuacán"},
		"15120": {"15", "Zumpango"},
		"15121": {"15", "Cuautitlán Izcalli"},
		"15122": {"15", "Valle de Chalco Solidaridad"},
		"15123": {"15", "Luvianos"},
		"15124": {"15", "San José del Rincón"},
		"15125": {"15", "Tonanitla"},
		"16001": {"16", "Acuitzio"},
		"16002": {"16", "Aguililla"},
		"16003": {"16", "Álvaro Obregón"},
		"16004": {"16", "Angamacutiro"},
		"16005": {"16", "Angangueo"},
		"16006": {"16", "Apatzingán"},
		"16007": {"16", "Aporo"},
		"16008": {"16", "Aquila"},
		"16009": {"16", "Ario"},
		"16010": {"16", "Arteaga"},
		"16011": {"16", "Briseñas"},
		"16012": {"16", "Buenavista"},
		"16013": {"16", "Carácuaro"},
		"16014": {"16", "Coahuayana"},
		"16015": {"16", "Coalcomán de Vázquez Pallares"},
		"16016": {"16", "Coeneo"},
		"16017": {"16", "Contepec"},
		"16018": {"16", "Copándaro"},
		"16019": {"16", "Cotija"},
		"16020": {"16", "Cuitzeo"},
		"16021": {"16", "Charapan"},
		"16022": {"16", "Charo"},
		"16023": {"16", "Chavinda"},
		"16024": {"16", "Cherán"},
		"16025": {"16", "Chilchota"},
		"16026": {"16", "Chinicuila"},
		"16027": {"16", "Chucándiro"},
		"16028": {"16", "Churintzio"},
		"16029": {"16", "Churumuco"},
		"16030": {"16", "Ecuandureo"},
		"16031": {"16", "Epitacio Huerta"},
		"16032": {"16", "Erongarícuaro"},
		"16033": {"16", "Gabriel Zamora"},
		"16034": {"16", "Hidalgo"},
		"16035": {"16", "La Huacana"},
		"16036": {"16", "Huandacareo"},
		"16037": {"16", "Huaniqueo"},
		"16038": {"16", "Huetamo"},
		"16039": {"16", "Huiramba"},
		"16040": {"16", "Indaparapeo"},
		"16041": {"16", "Irimbo"},
		"16042": {"16", "Ixtlán"},
		"16043": {"16", "Jacona"},
		"16044": {"16", "Jiménez"},
		"16045": {"16", "Jiquilpan"},
		"16046": {"16", "Juárez"},
		"16047": {"16", "Jungapeo"},
		"16048": {"16", "Lagunillas"},
		"16049": {"16", "Madero"},
		"16050": {"16", "Maravatío"},
		"16051": {"16", "Marcos Castellanos"},
		"16052": {"16", "Lázaro Cárdenas"},
		"16053": {"16", "Morelia"},
		"16054": {"16", "Morelos"},
		"16055": {"16", "Múgica"},
		"16056": {"16", "Nahuatzen"},
		"16057": {"16", "Nocupétaro"},
		"16058": {"16", "Nuevo Parangaricutiro"},
		"16059": {"16", "Nuevo Urecho"},
		"16060": {"16", "Numarán"},
		"16061": {"16", "Ocampo"},
		"16062": {"16", "Pajacuarán"},
		"16063": {"16", "Panindícuaro"},
		"16064": {"16", "Parácuaro"},
		"16065": {"16", "Paracho"},
		"16066": {"16", "Pátzcuaro"},
		"16067": {"16", "Penjamillo"},
		"16068": {"16", "Peribán"},
		"16069": {"16", "La Piedad"},
		"16070": {"16", "Purépero"},
		"16071": {"16", "Puruándiro"},
		"16072": {"16", "Queréndaro"},
		"16073": {"16", "Quiroga"},
		"16074": {"16", "Cojumatlán de Régules"},
		"16075": {"16", "Los Reyes"},
		"16076": {"16", "Sahuayo"},
		"16077": {"16", "San Lucas"},
		"16078": {"16", "Santa Ana Maya"},
		"16079": {"16", "Salvador Escalante"},
		"16080": {"16", "Senguio"},
		"16081": {"16", "Susupuato"},
		"16082": {"16", "Tacámbaro"},
		"16083": {"16", "Tancítaro"},
		"16084": {"16", "Tangamandapio"},
		"16085": {"16", "Tangancícuaro"},
		"16086": {"16", "Tanhuato"},
		"16087": {"16", "Taretan"},
		"16088": {"16", "Tarímbaro"},
		"16089": {"16", "Tepalcatepec"},
		"16090": {"16", "Tingambato"},
		"16091": {"16", "Tingüindín"},
		"16092": {"16", "Tiquicheo de Nicolás Romero"},
		"16093": {"16", "Tlalpujahua"},
		"16094": {"16", "Tlazazalca"},
		"16095": {"16", "Tocumbo"},
		"16096": {"16", "Tumbiscatío"},
		"16097": {"16", "Turicato"},
		"16098": {"16", "Tuxpan"},
		"16099": {"16", "Tuzantla"},
		"16100": {"16", "Tzintzuntzan"},
		"16101": {"16", "Tzitzio"},
		"16102": {"16", "Uruapan"},
		"16103": {"16", "Venustiano Carranza"},
		"16104": {"16", "Villamar"},
		"16105": {"16", "Vista Hermosa"},
		"16106": {"16", "Yurécuaro"},
		"16107": {"16", "Zacapu"},
		"16108": {"16", "Zamora"},
		"16109": {"16", "Zináparo"},
		"16110": {"16", "Zinapécuaro"},
		"16111": {"16", "Ziracuaretiro"},
		"16112": {"16", "Zitácuaro"},
		"16113": {"16", "José Sixto Verduzco"},
		"17001": {"17", "Amacuzac"},
		"17002": {"17", "Atlatlahucan"},
		"17003": {"17", "Axochiapan"},
		"17004": {"17", "Ayala"},
		"17005": {"17", "Coatlán del Río"},
		"17006": {"17", "Cuautla"},
		"17007": {"17", "Cuernavaca"},
		"17008": {"17", "Emiliano Zapata"},
		"17009": {"17", "Huitzilac"},
		"17010": {"17", "Jantetelco"},
		"17011": {"17", "Jiutepec"},
		"17012": {"17", "Jojutla"},
		"17013": {"17", "Jonacatepec"},
		"17014": {"17", "Mazatepec"},
		"17015": {"17", "Miacatlán"},
		"17016": {"17", "Ocuituco"},
		"17017": {"17", "Puente de Ixtla"},
		"17018": {"17", "Temixco"},
		"17019": {"17", "Tepalcingo"},
		"17020": {"17", "Tepoztlán"},
		"17021": {"17", "Tetecala"},
		"17022": {"17", "Tetela del Volcán"},
		"17023": {"17", "Tlalnepantla"},
		"17024": {"17", "Tlaltizapán"},
		"17025": {"17", "Tlaquiltenango"},
		"17026": {"17", "Tlayacapan"},
		"17027": {"17", "Totolapan"},
		"17028": {"17", "Xochitepec"},
		"17029": {"17", "Yautepec"},
		"17030": {"17", "Yecapixtla"},
		"17031": {"17", "Zacatepec"},
		"17032": {"17", "Zacualpan"},
		"17033": {"17", "Temoac"},
		"18001": {"18", "Acaponeta"},
		"18002": {"18", "Ahuacatlán"},
		"18003": {"18", "Amatlán de Cañas"},
		"18004": {"18", "Compostela"},
		"18005": {"18", "Huajicori"},
		"18006": {"18", "Ixtlán del Río"},
		"18007": {"18", "Jala"},
		"18008": {"18", "Xalisco"},
		"18009": {"18", "Del Nayar"},
		"18010": {"18", "Rosamorada"},
		"18011": {"18", "Ruíz"},
		"18012": {"18", "San Blas"},
		"18013": {"18", "San Pedro Lagunillas"},
		"18014": {"18", "Santa María del Oro"},
		"18015": {"18", "Santiago Ixcuintla"},
		"18016": {"18", "Tecuala"},
		"18017": {"18", "Tepic"},
		"18018": {"18", "Tuxpan"},
		"18019": {"18", "La Yesca"},
		"18020": {"18", "Bahía de Banderas"},
		"19001": {"19", "Abasolo"},
		"19002": {"19", "Agualeguas"},
		"19003": {"19", "Los Aldamas"},
		"19004": {"19", "Allende"},
		"19005": {"19", "Anáhuac"},
		"19006": {"19", "Apodaca"},
		"19007": {"19", "Aramberri"},
		"19008": {"19", "Bustamante"},
		"19009": {"19", "Cadereyta Jiménez"},
		"19010": {"19", "Carmen"},
		"19011": {"19", "Cerralvo"},
		"19012": {"19", "Ciénega de Flores"},
		"19013": {"19", "China"},
		"19014": {"19", "Dr. Arroyo"},
		"19015": {"19", "Dr. Coss"},
		"19016": {"19", "Dr. González"},
		"19017": {"19", "Galeana"},
		"19018": {"19", "García"},
		"19019": {"19", "San Pedro Garza García"},
		"19020": {"19", "Gral. Bravo"},
		"19021": {"19", "Gral. Escobedo"},
		"19022": {"19", "Gral. Terán"},
		"19023": {"19", "Gral. Treviño"},
		"19024": {"19", "Gral. Zaragoza"},
		"19025": {"19", "Gral. Zuazua"},
		"19026": {"19", "Guadalupe"},
		"19027": {"19", "Los Herreras"},
		"19028": {"19", "Higueras"},
		"19029": {"19", "Hualahuises"},
		"19030": {"19", "Iturbide"},
		"19031": {"19", "Juárez"},
		"19032": {"19", "Lampazos de Naranjo"},
		"19033": {"19", "Linares"},
		"19034": {"19", "Marín"},
		"19035": {"19", "Melchor Ocampo"},
		"19036": {"19", "Mier y Noriega"},
		"19037": {"19", "Mina"},
		"19038": {"19", "Montemorelos"},
		"19039": {"19", "Monterrey"},
		"19040": {"19", "Parás"},
		"19041": {"19", "Pesquería"},
		"19042": {"19", "Los Ramones"},
		"19043": {"19", "Rayones"},
		"19044": {"19", "Sabinas Hidalgo"},
		"19045": {"19", "Salinas Victoria"},
		"19046": {"19", "San Nicolás de los Garza"},
		"19047": {"19", "Hidalgo"},
		"19048": {"19", "Santa Catarina"},
		"19049": {"19", "Santiago"},
		"19050": {"19", "Vallecillo"},
		"19051": {"19", "Villaldama"},
		"20001": {"20", "Abejones"},
		"20002": {"20", "Acatlán de Pérez Figueroa"},
		"20003": {"20", "Asunción Cacalotepec"},
		"20004": {"20", "Asunción Cuyotepeji"},
		"20005": {"20", "Asunción Ixtaltepec"},
		"20006": {"20", "Asunción Nochixtlán"},
		"20007": {"20", "Asunción Ocotlán"},
		"20008": {"20", "Asunción Tlacolulita"},
		"20009": {"20", "Ayotzintepec"},
		"20010": {"20", "El Barrio de la Soledad"},
		"20011": {"20", "Calihualá"},
		"20012": {"20", "Candelaria Loxicha"},
		"20013": {"20", "Ciénega de Zimatlán"},
		"20014": {"20", "Ciudad Ixtepec"},
		"20015": {"20", "Coatecas Altas"},
		"20016": {"20", "Coicoyán de las Flores"},
		"20017": {"20", "La Compañía"},
		"20018": {"20", "Concepción Buenavista"},
		"20019": {"20", "Concepción Pápalo"},
		"20020": {"20", "Constancia del Rosario"},
		"20021": {"20", "Cosolapa"},
		"20022": {"20", "Cosoltepec"},
		"20023": {"20", "Cuilápam de Guerrero"},
		"20024": {"20", "Cuyamecalco Villa de Zaragoza"},
		"20025": {"20", "Chahuites"},
		"20026": {"20", "Chalcatongo de Hidalgo"},
		"20027": {"20", "Chiquihuitlán de Benito Juárez"},
		"20028": {"20", "Heroica Ciudad de Ejutla de Crespo"},
		"20029": {"20", "Eloxochitlán de Flores Magón"},
		"20030": {"20", "El Espinal"},
		"20031": {"20", "Tamazulápam del Espíritu Santo"},
		"20032": {"20", "Fresnillo de Trujano"},
		"20033": {"20", "Guadalupe Etla"},
		"20034": {"20", "Guadalupe de Ramírez"},
		"20035": {"20", "Guelatao de Juárez"},
		"20036": {"20", "Guevea de Humboldt"},
		"20037": {"20", "Mesones Hidalgo"},
		"20038": {"20", "Villa Hidalgo"},
		"20039": {"20", "Heroica Ciudad de Huajuapan de León"},
		"20040": {"20", "Huautepec"},
		"20041": {"20", "Huautla de Jiménez"},
		"20042": {"20", "Ixtlán de Juárez"},
		"20043": {"20", "Heroica Ciudad de Juchitán de Zaragoza"},
		"20044": {"20", "Loma Bonita"},
		"20045": {"20", "Magdalena Apasco"},
		"20046": {"20", "Magdalena Jaltepec"},
		"20047": {"20", "Santa Magdalena Jicotlán"},
		"20048": {"20", "Magdalena Mixtepec"},
		"20049": {"20", "Magdalena Ocotlán"},
		"20050": {"20", "Magdalena Peñasco"},
		"20051": {"20", "Magdalena Teitipac"},
		"20052": {"20", "Magdalena Tequisistlán"},
		"20053": {"20", "Magdalena Tlacotepec"},
		"20054": {"20", "Magdalena Zahuatlán"},
		"20055": {"20", "Mariscala de Juárez"},
		"20056": {"20", "Mártires de Tacubaya"},
		"20057": {"20", "Matías Romero Avendaño"},
		"20058": {"20", "Mazatlán Villa de Flores"},
		"20059": {"20", "Miahuatlán de Porfirio Díaz"},
		"20060": {"20", "Mixistlán de la Reforma"},
		"20061": {"20", "Monjas"},
		"20062": {"20", "Natividad"},
		"20063": {"20", "Nazareno Etla"},
		"20064": {"20", "Nejapa de Madero"},
		"20065": {"20", "Ixpantepec Nieves"},
		"20066": {"20", "Santiago Niltepec"},
		"20067": {"20", "Oaxaca de Juárez"},
		"20068": {"20", "Ocotlán de Morelos"},
		"20069": {"20", "La Pe"},
		"20070": {"20", "Pinotepa de Don Luis"},
		"20071": {"20", "Pluma Hidalgo"},
		"20072": {"20", "San José del Progreso"},
		"20073": {"20", "Putla Villa de Guerrero"},
		"20074": {"20", "Santa Catarina Quioquitani"},
		"20075": {"20", "Reforma de Pineda"},
		"20076": {"20", "La Reforma"},
		"20077": {"20", "Reyes Etla"},
		"20078": {"20", "Rojas de Cuauhtémoc"},
		"20079": {"20", "Salina Cruz"},
		"20080": {"20", "San Agustín Amatengo"},
		"20081": {"20", "San Agustín Atenango"},
		"20082": {"20", "San Agustín Chayuco"},
		"20083": {"20", "San Agustín de las Juntas"},
		"20084": {"20", "San Agustín Etla"},
		"20085": {"20", "San Agustín Loxicha"},
		"20086": {"20", "San Agustín Tlacotepec"},
		"20087": {"20", "San Agustín Yatareni"},
		"20088": {"20", "San Andrés Cabecera Nueva"},
		"20089": {"20", "San Andrés Dinicuiti"},
		"20090": {"20", "San Andrés Huaxpaltepec"},
		"20091": {"20", "San Andrés Huayápam"},
		"20092": {"20", "San Andrés Ixtlahuaca"},
		"20093": {"20", "San Andrés Lagunas"},
		"20094": {"20", "San Andrés Nuxiño"},
		"20095": {"20", "San Andrés Paxtlán"},
		"20096": {"20", "San Andrés Sinaxtla"},
		"20097": {"20", "San Andrés Solaga"},
		"20098": {"20", "San Andrés Teotilálpam"},
		"20099": {"20", "San Andrés Tepetlapa"},
		"20100": {"20", "San Andrés Yaá"},
		"20101": {"20", "San Andrés Zabache"},
		"20102": {"20", "San Andrés Zautla"},
		"20103": {"20", "San Antonino Castillo Velasco"},
		"20104": {"20", "San Antonino el Alto"},
		"20105": {"20", "San Antonino Monte Verde"},
		"20106": {"20", "San Antonio Acutla"},
		"20107": {"20", "San Antonio de la Cal"},
		"20108": {"20", "San Antonio Huitepec"},
		"20109": {"20", "San Antonio Nanahuatípam"},
		"20110": {"20", "San Antonio Sinicahua"},
		"20111": {"20", "San Antonio Tepetlapa"},
		"20112": {"20", "San Baltazar Chichicápam"},
		"20113": {"20", "San Baltazar Loxicha"},
		"20114": {"20", "San Baltazar Yatzachi el Bajo"},
		"20115": {"20", "San Bartolo Coyotepec"},
		"20116": {"20", "San Bartolomé Ayautla"},
		"20117": {"20", "San Bartolomé Loxicha"},
		"20118": {"20", "San Bartolomé Quialana"},
		"20119": {"20", "San Bartolomé Yucuañe"},
		"20120": {"20", "San Bartolomé Zoogocho"},
		"20121": {"20", "San Bartolo Soyaltepec"},
		"20122": {"20", "San Bartolo Yautepec"},
		"20123": {"20", "San Bernardo Mixtepec"},
		"20124": {"20", "San Blas Atempa"},
		"20125": {"20", "San Carlos Yautepec"},
		"20126": {"20", "San Cristóbal Amatlán"},
		"20127": {"20", "San Cristóbal Amoltepec"},
		"20128": {"20", "San Cristóbal Lachirioag"},
		"20129": {"20", "San Cristóbal Suchixtlahuaca"},
		"20130": {"20", "San Dionisio del Mar"},
		"20131": {"20", "San Dionisio Ocotepec"},
		"20132": {"20", "San Dionisio Ocotlán"},
		"20133": {"20", "San Esteban Atatlahuca"},
		"20134": {"20", "San Felipe Jalapa de Díaz"},
		"20135": {"20", "San Felipe Tejalápam"},
		"20136": {"20", "San Felipe Usila"},
		"20137": {"20", "San Francisco Cahuacuá"},
		"20138": {"20", "San Francisco Cajonos"},
		"20139": {"20", "San Francisco Chapulapa"},
		"20140": {"20", "San Francisco Chindúa"},
		"20141": {"20", "San Francisco del Mar"},
		"20142": {"20", "San Francisco Huehuetlán"},
		"20143": {"20", "San Francisco Ixhuatán"},
		"20144": {"20", "San Francisco Jaltepetongo"},
		"20145": {"20", "San Francisco Lachigoló"},
		"20146": {"20", "San Francisco Logueche"},
		"20147": {"20", "San Francisco Nuxaño"},
		"20148": {"20", "San Francisco Ozolotepec"},
		"20149": {"20", "San Francisco Sola"},
		"20150": {"20", "San Francisco Telixtlahuaca"},
		"20151": {"20", "San Francisco Teopan"},
		"20152": {"20", "San Francisco Tlapancingo"},
		"20153": {"20", "San Gabriel Mixtepec"},
		"20154": {"20", "San Ildefonso Amatlán"},
		"20155": {"20", "San Ildefonso Sola"},
		"20156": {"20", "San Ildefonso Villa Alta"},
		"20157": {"20", "San Jacinto Amilpas"},
		"20158": {"20", "San Jacinto Tlacotepec"},
		"20159": {"20", "San Jerónimo Coatlán"},
		"20160": {"20", "San Jerónimo Silacayoapilla"},
		"20161": {"20", "San Jerónimo Sosola"},
		"20162": {"20", "San Jerónimo Taviche"},
		"20163": {"20", "San Jerónimo Tecóatl"},
		"20164": {"20", "San Jorge Nuchita"},
		"20165": {"20", "San José Ayuquila"},
		"20166": {"20", "San José Chiltepec"},
		"20167": {"20", "San José del Peñasco"},
		"20168": {"20", "San José Estancia Grande"},
		"20169": {"20", "San José Independencia"},
		"20170": {"20", "San José Lachiguiri"},
		"20171": {"20", "San José Tenango"},
		"20172": {"20", "San Juan Achiutla"},
		"20173": {"20", "San Juan Atepec"},
		"20174": {"20", "Ánimas Trujano"},
		"20175": {"20", "San Juan Bautista Atatlahuca"},
		"20176": {"20", "San Juan Bautista Coixtlahuaca"},
		"20177": {"20", "San Juan Bautista Cuicatlán"},
		"20178": {"20", "San Juan Bautista Guelache"},
		"20179": {"20", "San Juan Bautista Jayacatlán"},
		"20180": {"20", "San Juan Bautista Lo de Soto"},
		"20181": {"20", "San Juan Bautista Suchitepec"},
		"20182": {"20", "San Juan Bautista Tlacoatzintepec"},
		"20183": {"20", "San Juan Bautista Tlachichilco"},
		"20184": {"20", "San Juan Bautista Tuxtepec"},
		"20185": {"20", "San Juan Cacahuatepec"},
		"20186": {"20", "San Juan Cieneguilla"},
		"20187": {"20", "San Juan Coatzóspam"},
		"20188": {"20", "San Juan Colorado"},
		"20189": {"20", "San Juan Comaltepec"},
		"20190": {"20", "San Juan Cotzocón"},
		"20191": {"20", "San Juan Chicomezúchil"},
		"20192": {"20", "San Juan Chilateca"},
		"20193": {"20", "San Juan del Estado"},
		"20194": {"20", "San Juan del Río"},
		"20195": {"20", "San Juan Diuxi"},
		"20196": {"20", "San Juan Evangelista Analco"},
		"20197": {"20", "San Juan Guelavía"},
		"20198": {"20", "San Juan Guichicovi"},
		"20199": {"20", "San Juan Ihualtepec"},
		"20200": {"20", "San Juan Juquila Mixes"},
		"20201": {"20", "San Juan Juquila Vijanos"},
		"20202": {"20", "San Juan Lachao"},
		"20203": {"20", "San Juan Lachigalla"},
		"20204": {"20", "San Juan Lajarcia"},
		"20205": {"20", "San Juan Lalana"},
		"20206": {"20", "San Juan de los Cués"},
		"20207": {"20", "San Juan Mazatlán"},
		"20208": {"20", "San Juan Mixtepec -Dto. 08 -"},
		"20209": {"20", "San Juan Mixtepec -Dto. 26 -"},
		"20210": {"20", "San Juan Ñumí"},
		"20211": {"20", "San Juan Ozolotepec"},
		"20212": {"20", "San Juan Petlapa"},
		"20213": {"20", "San Juan Quiahije"},
		"20214": {"20", "San Juan Quiotepec"},
		"20215": {"20", "San Juan Sayultepec"},
		"20216": {"20", "San Juan Tabaá"},
		"20217": {"20", "San Juan Tamazola"},
		"20218": {"20", "San Juan Teita"},
		"20219": {"20", "San Juan Teitipac"},
		"20220": {"20", "San Juan Tepeuxila"},
		"20221": {"20", "San Juan Teposcolula"},
		"20222": {"20", "San Juan Yaeé"},
		"20223": {"20", "San Juan Yatzona"},
		"20224": {"20", "San Juan Yucuita"},
		"20225": {"20", "San Lorenzo"},
		"20226": {"20", "San Lorenzo Albarradas"},
		"20227": {"20", "San Lorenzo Cacaotepec"},
		"20228": {"20", "San Lorenzo Cuaunecuiltitla"},
		"20229": {"20", "San Lorenzo Texmelúcan"},
		"20230": {"20", "San Lorenzo Victoria"},
		"20231": {"20", "San Lucas Camotlán"},
		"20232": {"20", "San Lucas Ojitlán"},
		"20233": {"20", "San Lucas Quiaviní"},
		"20234": {"20", "San Lucas Zoquiápam"},
		"20235": {"20", "San Luis Amatlán"},
		"20236": {"20", "San Marcial Ozolotepec"},
		"20237": {"20", "San Marcos Arteaga"},
		"20238": {"20", "San Martín de los Cansecos"},
		"20239": {"20", "San Martín Huamelúlpam"},
		"20240": {"20", "San Martín Itunyoso"},
		"20241": {"20", "San Martín Lachilá"},
		"20242": {"20", "San Martín Peras"},
		"20243": {"20", "San Martín Tilcajete"},
		"20244": {"20", "San Martín Toxpalan"},
		"20245": {"20", "San Martín Zacatepec"},
		"20246": {"20", "San Mateo Cajonos"},
		"20247": {"20", "Capulálpam de Méndez"},
		"20248": {"20", "San Mateo del Mar"},
		"20249": {"20", "San Mateo Yoloxochitlán"},
		"20250": {"20", "San Mateo Etlatongo"},
		"20251": {"20", "San Mateo Nejápam"},
		"20252": {"20", "San Mateo Peñasco"},
		"20253": {"20", "San Mateo Piñas"},
		"20254": {"20", "San Mateo Río Hondo"},
		"20255": {"20", "San Mateo Sindihui"},
		"20256": {"20", "San Mateo Tlapiltepec"},
		"20257": {"20", "San Melchor Betaza"},
		"20258": {"20", "San Miguel Achiutla"},
		"20259": {"20", "San Miguel Ahuehuetitlán"},
		"20260": {"20", "San Miguel Aloápam"},
		"20261": {"20", "San Miguel Amatitlán"},
		"20262": {"20", "San Miguel Amatlán"},
		"20263": {"20", "San Miguel Coatlán"},
		"20264": {"20", "San Miguel Chicahua"},
		"20265": {"20", "San Miguel Chimalapa"},
		"20266": {"20", "San Miguel del Puerto"},
		"20267": {"20", "San Miguel del Río"},
		"20268": {"20", "San Miguel Ejutla"},
		"20269": {"20", "San Miguel el Grande"},
		"20270": {"20", "San Miguel Huautla"},
		"20271": {"20", "San Miguel Mixtepec"},
		"20272": {"20", "San Miguel Panixtlahuaca"},
		"20273": {"20", "San Miguel Peras"},
		"20274": {"20", "San Miguel Piedras"},
		"20275": {"20", "San Miguel Quetzaltepec"},
		"20276": {"20", "San Miguel Santa Flor"},
		"20277": {"20", "Villa Sola de Vega"},
		"20278": {"20", "San Miguel Soyaltepec"},
		"20279": {"20", "San Miguel Suchixtepec"},
		"20280": {"20", "Villa Talea de Castro"},
		"20281": {"20", "San Miguel Tecomatlán"},
		"20282": {"20", "San Miguel Tenango"},
		"20283": {"20", "San Miguel Tequixtepec"},
		"20284": {"20", "San Miguel Tilquiápam"},
		"20285": {"20", "San Miguel Tlacamama"},
		"20286": {"20", "San Miguel Tlacotepec"},
		"20287": {"20", "San Miguel Tulancingo"},
		"20288": {"20", "San Miguel Yotao"},
		"20289": {"20", "San Nicolás"},
		"20290": {"20", "San Nicolás Hidalgo"},
		"20291": {"20", "San Pablo Coatlán"},
		"20292": {"20", "San Pablo Cuatro Venados"},
		"20293": {"20", "San Pablo Etla"},
		"20294": {"20", "San Pablo Huitzo"},
		"20295": {"20", "San Pablo Huixtepec"},
		"20296": {"20", "San Pablo Macuiltianguis"},
		"20297": {"20", "San Pablo Tijaltepec"},
		"20298": {"20", "San Pablo Villa de Mitla"},
		"20299": {"20", "San Pablo Yaganiza"},
		"20300": {"20", "San Pedro Amuzgos"},
		"20301": {"20", "San Pedro Apóstol"},
		"20302": {"20", "San Pedro Atoyac"},
		"20303": {"20", "San Pedro Cajonos"},
		"20304": {"20", "San Pedro Coxcaltepec Cántaros"},
		"20305": {"20", "San Pedro Comitancillo"},
		"20306": {"20", "San Pedro el Alto"},
		"20307": {"20", "San Pedro Huamelula"},
		"20308": {"20", "San Pedro Huilotepec"},
		"20309": {"20", "San Pedro Ixcatlán"},
		"20310": {"20", "San Pedro Ixtlahuaca"},
		"20311": {"20", "San Pedro Jaltepetongo"},
		"20312": {"20", "San Pedro Jicayán"},
		"20313": {"20", "San Pedro Jocotipac"},
		"20314": {"20", "San Pedro Juchatengo"},
		"20315": {"20", "San Pedro Mártir"},
		"20316": {"20", "San Pedro Mártir Quiechapa"},
		"20317": {"20", "San Pedro Mártir Yucuxaco"},
		"20318": {"20", "San Pedro Mixtepec -Dto. 22 -"},
		"20319": {"20", "San Pedro Mixtepec -Dto. 26 -"},
		"20320": {"20", "San Pedro Molinos"},
		"20321": {"20", "San Pedro Nopala"},
		"20322": {"20", "San Pedro Ocopetatillo"},
		"20323": {"20", "San Pedro Ocotepec"},
		"20324": {"20", "San Pedro Pochutla"},
		"20325": {"20", "San Pedro Quiatoni"},
		"20326": {"20", "San Pedro Sochiápam"},
		"20327": {"20", "San Pedro Tapanatepec"},
		"20328": {"20", "San Pedro Taviche"},
		"20329": {"20", "San Pedro Teozacoalco"},
		"20330": {"20", "San Pedro Teutila"},
		"20331": {"20", "San Pedro Tidaá"},
		"20332": {"20", "San Pedro Topiltepec"},
		"20333": {"20", "San Pedro Totolápam"},
		"20334": {"20", "Villa de Tututepec de Melchor Ocampo"},
		"20335": {"20", "San Pedro Yaneri"},
		"20336": {"20", "San Pedro Yólox"},
		"20337": {"20", "San Pedro y San Pablo Ayutla"},
		"20338": {"20", "Villa de Etla"},
		"20339": {"20", "San Pedro y San Pablo Teposcolula"},
		"20340": {"20", "San Pedro y San Pablo Tequixtepec"},
		"20341": {"20", "San Pedro Yucunama"},
		"20342": {"20", "San Raymundo Jalpan"},
		"20343": {"20", "San Sebastián Abasolo"},
		"20344": {"20", "San Sebastián Coatlán"},
		"20345": {"20", "San Sebastián Ixcapa"},
		"20346": {"20", "San Sebastián Nicananduta"},
		"20347": {"20", "San Sebastián Río Hondo"},
		"20348": {"20", "San Sebastián Tecomaxtlahuaca"},
		"20349": {"20", "San Sebastián Teitipac"},
		"20350": {"20", "San Sebastián Tutla"},
		"20351": {"20", "San Simón Almolongas"},
		"20352": {"20", "San Simón Zahuatlán"},
		"20353": {"20", "Santa Ana"},
		"20354": {"20", "Santa Ana Ateixtlahuaca"},
		"20355": {"20", "Santa Ana Cuauhtémoc"},
		"20356": {"20", "Santa Ana del Valle"},
		"20357": {"20", "Santa Ana Tavela"},
		"20358": {"20", "Santa Ana Tlapacoyan"},
		"20359": {"20", "Santa Ana Yareni"},
		"20360": {"20", "Santa Ana Zegache"},
		"20361": {"20", "Santa Catalina Quierí"},
		"20362": {"20", "Santa Catarina Cuixtla"},
		"20363": {"20", "Santa Catarina Ixtepeji"},
		"20364": {"20", "Santa Catarina Juquila"},
		"20365": {"20", "Santa Catarina Lachatao"},
		"20366": {"20", "Santa Catarina Loxicha"},
		"20367": {"20", "Santa Catarina Mechoacán"},
		"20368": {"20", "Santa Catarina Minas"},
		"20369": {"20", "Santa Catarina Quiané"},
		"20370": {"20", "Santa Catarina Tayata"},
		"20371": {"20", "Santa Catarina Ticuá"},
		"20372": {"20", "Santa Catarina Yosonotú"},
		"20373": {"20", "Santa Catarina Zapoquila"},
		"20374": {"20", "Santa Cruz Acatepec"},
		"20375": {"20", "Santa Cruz Amilpas"},
		"20376": {"20", "Santa Cruz de Bravo"},
		"20377": {"20", "Santa Cruz Itundujia"},
		"20378": {"20", "Santa Cruz Mixtepec"},
		"20379": {"20", "Santa Cruz Nundaco"},
		"20380": {"20", "Santa Cruz Papalutla"},
		"20381": {"20", "Santa Cruz Tacache de Mina"},
		"20382": {"20", "Santa Cruz Tacahua"},
		"20383": {"20", "Santa Cruz Tayata"},
		"20384": {"20", "Santa Cruz Xitla"},
		"20385": {"20", "Santa Cruz Xoxocotlán"},
		"20386": {"20", "Santa Cruz Zenzontepec"},
		"20387": {"20", "Santa Gertrudis"},
		"20388": {"20", "Santa Inés del Monte"},
		"20389": {"20", "Santa Inés Yatzeche"},
		"20390": {"20", "Santa Lucía del Camino"},
		"20391": {"20", "Santa Lucía Miahuatlán"},
		"20392": {"20", "Santa Lucía Monteverde"},
		"20393": {"20", "Santa Lucía Ocotlán"},
		"20394": {"20", "Santa María Alotepec"},
		"20395": {"20", "Santa María Apazco"},
		"20396": {"20", "Santa María la Asunción"},
		"20397": {"20", "Heroica Ciudad de Tlaxiaco"},
		"20398": {"20", "Ayoquezco de Aldama"},
		"20399": {"20", "Santa María Atzompa"},
		"20400": {"20", "Santa María Camotlán"},
		"20401": {"20", "Santa María Colotepec"},
		"20402": {"20", "Santa María Cortijo"},
		"20403": {"20", "Santa María Coyotepec"},
		"20404": {"20", "Santa María Chachoápam"},
		"20405": {"20", "Villa de Chilapa de Díaz"},
		"20406": {"20", "Santa María Chilchotla"},
		"20407": {"20", "Santa María Chimalapa"},
		"20408": {"20", "Santa María del Rosario"},
		"20409": {"20", "Santa María del Tule"},
		"20410": {"20", "Santa María Ecatepec"},
		"20411": {"20", "Santa María Guelacé"},
		"20412": {"20", "Santa María Guienagati"},
		"20413": {"20", "Santa María Huatulco"},
		"20414": {"20", "Santa María Huazolotitlán"},
		"20415": {"20", "Santa María Ipalapa"},
		"20416": {"20", "Santa María Ixcatlán"},
		"20417": {"20", "Santa María Jacatepec"},
		"20418": {"20", "Santa María Jalapa del Marqués"},
		"20419": {"20", "Santa María Jaltianguis"},
		"20420": {"20", "Santa María Lachixío"},
		"20421": {"20", "Santa María Mixtequilla"},
		"20422": {"20", "Santa María Nativitas"},
		"20423": {"20", "Santa María Nduayaco"},
		"20424": {"20", "Santa María Ozolotepec"},
		"20425": {"20", "Santa María Pápalo"},
		"20426": {"20", "Santa María Peñoles"},
		"20427": {"20", "Santa María Petapa"},
		"20428": {"20", "Santa María Quiegolani"},
		"20429": {"20", "Santa María Sola"},
		"20430": {"20", "Santa María Tataltepec"},
		"20431": {"20", "Santa María Tecomavaca"},
		"20432": {"20", "Santa María Temaxcalapa"},
		"20433": {"20", "Santa María Temaxcaltepec"},
		"20434": {"20", "Santa María Teopoxco"},
		"20435": {"20", "Santa María Tepantlali"},
		"20436": {"20", "Santa María Texcatitlán"},
		"20437": {"20", "Santa María Tlahuitoltepec"},
		"20438": {"20", "Santa María Tlalixtac"},
		"20439": {"20", "Santa María Tonameca"},
		"20440": {"20", "Santa María Totolapilla"},
		"20441": {"20", "Santa María Xadani"},
		"20442": {"20", "Santa María Yalina"},
		"20443": {"20", "Santa María Yavesía"},
		"20444": {"20", "Santa María Yolotepec"},
		"20445": {"20", "Santa María Yosoyúa"},
		"20446": {"20", "Santa María Yucuhiti"},
		"20447": {"20", "Santa María Zacatepec"},
		"20448": {"20", "Santa María Zaniza"},
		"20449": {"20", "Santa María Zoquitlán"},
		"20450": {"20", "Santiago Amoltepec"},
		"20451": {"20", "Santiago Apoala"},
		"20452": {"20", "Santiago Apóstol"},
		"20453": {"20", "Santiago Astata"},
		"20454": {"20", "Santiago Atitlán"},
		"20455": {"20", "Santiago Ayuquililla"},
		"20456": {"20", "Santiago Cacaloxtepec"},
		"20457": {"20", "Santiago Camotlán"},
		"20458": {"20", "Santiago Comaltepec"},
		"20459": {"20", "Santiago Chazumba"},
		"20460": {"20", "Santiago Choápam"},
		"20461": {"20", "Santiago del Río"},
		"20462": {"20", "Santiago Huajolotitlán"},
		"20463": {"20", "Santiago Huauclilla"},
		"20464": {"20", "Santiago Ihuitlán Plumas"},
		"20465": {"20", "Santiago Ixcuintepec"},
		"20466": {"20", "Santiago Ixtayutla"},
		"20467": {"20", "Santiago Jamiltepec"},
		"20468": {"20", "Santiago Jocotepec"},
		"20469": {"20", "Santiago Juxtlahuaca"},
		"20470": {"20", "Santiago Lachiguiri"},
		"20471": {"20", "Santiago Lalopa"},
		"20472": {"20", "Santiago Laollaga"},
		"20473": {"20", "Santiago Laxopa"},
		"20474": {"20", "Santiago Llano Grande"},
		"20475": {"20", "Santiago Matatlán"},
		"20476": {"20", "Santiago Miltepec"},
		"20477": {"20", "Santiago Minas"},
		"20478": {"20", "Santiago Nacaltepec"},
		"20479": {"20", "Santiago Nejapilla"},
		"20480": {"20", "Santiago Nundiche"},
		"20481": {"20", "Santiago Nuyoó"},
		"20482": {"20", "Santiago Pinotepa Nacional"},
		"20483": {"20", "Santiago Suchilquitongo"},
		"20484": {"20", "Santiago Tamazola"},
		"20485": {"20", "Santiago Tapextla"},
		"20486": {"20", "Villa Tejúpam de la Unión"},
		"20487": {"20", "Santiago Tenango"},
		"20488": {"20", "Santiago Tepetlapa"},
		"20489": {"20", "Santiago Tetepec"},
		"20490": {"20", "Santiago Texcalcingo"},
		"20491": {"20", "Santiago Textitlán"},
		"20492": {"20", "Santiago Tilantongo"},
		"20493": {"20", "Santiago Tillo"},
		"20494": {"20", "Santiago Tlazoyaltepec"},
		"20495": {"20", "Santiago Xanica"},
		"20496": {"20", "Santiago Xiacuí"},
		"20497": {"20", "Santiago Yaitepec"},
		"20498": {"20", "Santiago Yaveo"},
		"20499": {"20", "Santiago Yolomécatl"},
		"20500": {"20", "Santiago Yosondúa"},
		"20501": {"20", "Santiago Yucuyachi"},
		"20502": {"20", "Santiago Zacatepec"},
		"20503": {"20", "Santiago Zoochila"},
		"20504": {"20", "Nuevo Zoquiápam"},
		"20505": {"20", "Santo Domingo Ingenio"},
		"20506": {"20", "Santo Domingo Albarradas"},
		"20507": {"20", "Santo Domingo Armenta"},
		"20508": {"20", "Santo Domingo Chihuitán"},
		"20509": {"20", "Santo Domingo de Morelos"},
		"20510": {"20", "Santo Domingo Ixcatlán"},
		"20511": {"20", "Santo Domingo Nuxaá"},
		"20512": {"20", "Santo Domingo Ozolotepec"},
		"20513": {"20", "Santo Domingo Petapa"},
		"20514": {"20", "Santo Domingo Roayaga"},
		"20515": {"20", "Santo Domingo Tehuantepec"},
		"20516": {"20", "Santo Domingo Teojomulco"},
		"20517": {"20", "Santo Domingo Tepuxtepec"},
		"20518": {"20", "Santo Domingo Tlatayápam"},
		"20519": {"20", "Santo Domingo Tomaltepec"},
		"20520": {"20", "Santo Domingo Tonalá"},
		"20521": {"20", "Santo Domingo Tonaltepec"},
		"20522": {"20", "Santo Domingo Xagacía"},
		"20523": {"20", "Santo Domingo Yanhuitlán"},
		"20524": {"20", "Santo Domingo Yodohino"},
		"20525": {"20", "Santo Domingo Zanatepec"},
		"20526": {"20", "Santos Reyes Nopala"},
		"20527": {"20", "Santos Reyes Pápalo"},
		"20528": {"20", "Santos Reyes Tepejillo"},
		"20529": {"20", "Santos Reyes Yucuná"},
		"20530": {"20", "Santo Tomás Jalieza"},
		"20531": {"20", "Santo Tomás Mazaltepec"},
		"20532": {"20", "Santo Tomás Ocotepec"},
		"20533": {"20", "Santo Tomás Tamazulapan"},
		"20534": {"20", "San Vicente Coatlán"},
		"20535": {"20", "San Vicente Lachixío"},
		"20536": {"20", "San Vicente Nuñú"},
		"20537": {"20", "Silacayoápam"},
		"20538": {"20", "Sitio de Xitlapehua"},
		"20539": {"20", "Soledad Etla"},
		"20540": {"20", "Villa de Tamazulápam del Progreso"},
		"20541": {"20", "Tanetze de Zaragoza"},
		"20542": {"20", "Taniche"},
		"20543": {"20", "Tataltepec de Valdés"},
		"20544": {"20", "Teococuilco de Marcos Pérez"},
		"20545": {"20", "Teotitlán de Flores Magón"},
		"20546": {"20", "Teotitlán del Valle"},
		"20547": {"20", "Teotongo"},
		"20548": {"20", "Tepelmeme Villa de Morelos"},
		"20549": {"20", "Tezoatlán de Segura y Luna"},
		"20550": {"20", "San Jerónimo Tlacochahuaya"},
		"20551": {"20", "Tlacolula de Matamoros"},
		"20552": {"20", "Tlacotepec Plumas"},
		"20553": {"20", "Tlalixtac de Cabrera"},
		"20554": {"20", "Totontepec Villa de Morelos"},
		"20555": {"20", "Trinidad Zaachila"},
		"20556": {"20", "La Trinidad Vista Hermosa"},
		"20557": {"20", "Unión Hidalgo"},
		"20558": {"20", "Valerio Trujano"},
		"20559": {"20", "San Juan Bautista Valle Nacional"},
		"20560": {"20", "Villa Díaz Ordaz"},
		"20561": {"20", "Yaxe"},
		"20562": {"20", "Magdalena Yodocono de Porfirio Díaz"},
		"20563": {"20", "Yogana"},
		"20564": {"20", "Yutanduchi de Guerrero"},
		"20565": {"20", "Villa de Zaachila"},
		"20566": {"20", "San Mateo Yucutindó"},
		"20567": {"20", "Zapotitlán Lagunas"},
		"20568": {"20", "Zapotitlán Palmas"},
		"20569": {"20", "Santa Inés de Zaragoza"},
		"20570": {"20", "Zimatlán de Álvarez"},
		"21001": {"21", "Acajete"},
		"21002": {"21", "Acateno"},
		"21003": {"21", "Acatlán"},
		"21004": {"21", "Acatzingo"},
		"21005": {"21", "Acteopan"},
		"21006": {"21", "Ahuacatlán"},
		"21007": {"21", "Ahuatlán"},
		"21008": {"21", "Ahuazotepec"},
		"21009": {"21", "Ahuehuetitla"},
		"21010": {"21", "Ajalpan"},
		"21011": {"21", "Albino Zertuche"},
		"21012": {"21", "Aljojuca"},
		"21013": {"21", "Altepexi"},
		"21014": {"21", "Amixtlán"},
		"21015": {"21", "Amozoc"},
		"21016": {"21", "Aquixtla"},
		"21017": {"21", "Atempan"},
		"21018": {"21", "Atexcal"},
		"21019": {"21", "Atlixco"},
		"21020": {"21", "Atoyatempan"},
		"21021": {"21", "Atzala"},
		"21022": {"21", "Atzitzihuacán"},
		"21023": {"21", "Atzitzintla"},
		"21024": {"21", "Axutla"},
		"21025": {"21", "Ayotoxco de Guerrero"},
		"21026": {"21", "Calpan"},
		"21027": {"21", "Caltepec"},
		"21028": {"21", "Camocuautla"},
		"21029": {"21", "Caxhuacan"},
		"21030": {"21", "Coatepec"},
		"21031": {"21", "Coatzingo"},
		"21032": {"21", "Cohetzala"},
		"21033": {"21", "Cohuecan"},
		"21034": {"21", "Coronango"},
		"21035": {"21", "Coxcatlán"},
		"21036": {"21", "Coyomeapan"},
		"21037": {"21", "Coyotepec"},
		"21038": {"21", "Cuapiaxtla de Madero"},
		"21039": {"21", "Cuautempan"},
		"21040": {"21", "Cuautinchán"},
		"21041": {"21", "Cuautlancingo"},
		"21042": {"21", "Cuayuca de Andrade"},
		"21043": {"21", "Cuetzalan del Progreso"},
		"21044": {"21", "Cuyoaco"},
		"21045": {"21", "Chalchicomula de Sesma"},
		"21046": {"21", "Chapulco"},
		"21047": {"21", "Chiautla"},
		"21048": {"21", "Chiautzingo"},
		"21049": {"21", "Chiconcuautla"},
		"21050": {"21", "Chichiquila"},
		"21051": {"21", "Chietla"},
		"21052": {"21", "Chigmecatitlán"},
		"21053": {"21", "Chignahuapan"},
		"21054": {"21", "Chignautla"},
		"21055": {"21", "Chila"},
		"21056": {"21", "Chila de la Sal"},
		"21057": {"21", "Honey"},
		"21058": {"21", "Chilchotla"},
		"21059": {"21", "Chinantla"},
		"21060": {"21", "Domingo Arenas"},
		"21061": {"21", "Eloxochitlán"},
		"21062": {"21", "Epatlán"},
		"21063": {"21", "Esperanza"},
		"21064": {"21", "Francisco Z. Mena"},
		"21065": {"21", "General Felipe Ángeles"},
		"21066": {"21", "Guadalupe"},
		"21067": {"21", "Guadalupe Victoria"},
		"21068": {"21", "Hermenegildo Galeana"},
		"21069": {"21", "Huaquechula"},
		"21070": {"21", "Huatlatlauca"},
		"21071": {"21", "Huauchinango"},
		"21072": {"21", "Huehuetla"},
		"21073": {"21", "Huehuetlán el Chico"},
		"21074": {"21", "Huejotzingo"},
		"21075": {"21", "Hueyapan"},
		"21076": {"21", "Hueytamalco"},
		"21077": {"21", "Hueytlalpan"},
		"21078": {"21", "Huitzilan de Serdán"},
		"21079": {"21", "Huitziltepec"},
		"21080": {"21", "Atlequizayan"},
		"21081": {"21", "Ixcamilpa de Guerrero"},
		"21082": {"21", "Ixcaquixtla"},
		"21083": {"21", "Ixtacamaxtitlán"},
		"21084": {"21", "Ixtepec"},
		"21085": {"21", "Izúcar de Matamoros"},
		"21086": {"21", "Jalpan"},
		"21087": {"21", "Jolalpan"},
		"21088": {"21", "Jonotla"},
		"21089": {"21", "Jopala"},
		"21090": {"21", "Juan C. Bonilla"},
		"21091": {"21", "Juan Galindo"},
		"21092": {"21", "Juan N. Méndez"},
		"21093": {"21", "Lafragua"},
		"21094": {"21", "Libres"},
		"21095": {"21", "La Magdalena Tlatlauquitepec"},
		"21096": {"21", "Mazapiltepec de Juárez"},
		"21097": {"21", "Mixtla"},
		"21098": {"21", "Molcaxac"},
		"21099": {"21", "Cañada Morelos"},
		"21100": {"21", "Naupan"},
		"21101": {"21", "Nauzontla"},
		"21102": {"21", "Nealtican"},
		"21103": {"21", "Nicolás Bravo"},
		"21104": {"21", "Nopalucan"},
		"21105": {"21", "Ocotepec"},
		"21106": {"21", "Ocoyucan"},
		"21107": {"21", "Olintla"},
		"21108": {"21", "Oriental"},
		"21109": {"21", "Pahuatlán"},
		"21110": {"21", "Palmar de Bravo"},
		"21111": {"21", "Pantepec"},
		"21112": {"21", "Petlalcingo"},
		"21113": {"21", "Piaxtla"},
		"21114": {"21", "Puebla"},
		"21115": {"21", "Quecholac"},
		"21116": {"21", "Quimixtlán"},
		"21117": {"21", "Rafael Lara Grajales"},
		"21118": {"21", "Los Reyes de Juárez"},
		"21119": {"21", "San Andrés Cholula"},
		"21120": {"21", "San Antonio Cañada"},
		"21121": {"21", "San Diego la Mesa Tochimiltzingo"},
		"21122": {"21", "San Felipe Teotlalcingo"},
		"21123": {"21", "San Felipe Tepatlán"},
		"21124": {"21", "San Gabriel Chilac"},
		"21125": {"21", "San Gregorio Atzompa"},
		"21126": {"21", "San Jerónimo Tecuanipan"},
		"21127": {"21", "San Jerónimo Xayacatlán"},
		"21128": {"21", "San José Chiapa"},
		"21129": {"21", "San José Miahuatlán"},
		"21130": {"21", "San Juan Atenco"},
		"21131": {"21", "San Juan Atzompa"},
		"21132": {"21", "San Martín Texmelucan"},
		"21133": {"21", "San Martín Totoltepec"},
		"21134": {"21", "San Matías Tlalancaleca"},
		"21135": {"21", "San Miguel Ixitlán"},
		"21136": {"21", "San Miguel Xoxtla"},
		"21137": {"21", "San Nicolás Buenos Aires"},
		"21138": {"21", "San Nicolás de los Ranchos"},
		"21139": {"21", "San Pablo Anicano"},
		"21140": {"21", "San Pedro Cholula"},
		"21141": {"21", "San Pedro Yeloixtlahuaca"},
		"21142": {"21", "San Salvador el Seco"},
		"21143": {"21", "San Salvador el Verde"},
		"21144": {"21", "San Salvador Huixcolotla"},
		"21145": {"21", "San Sebastián Tlacotepec"},
		"21146": {"21", "Santa Catarina Tlaltempan"},
		"21147": {"21", "Santa Inés Ahuatempan"},
		"21148": {"21", "Santa Isabel Cholula"},
		"21149": {"21", "Santiago Miahuatlán"},
		"21150": {"21", "Huehuetlán el Grande"},
		"21151": {"21", "Santo Tomás Hueyotlipan"},
		"21152": {"21", "Soltepec"},
		"21153": {"21", "Tecali de Herrera"},
		"21154": {"21", "Tecamachalco"},
		"21155": {"21", "Tecomatlán"},
		"21156": {"21", "Tehuacán"},
		"21157": {"21", "Tehuitzingo"},
		"21158": {"21", "Tenampulco"},
		"21159": {"21", "Teopantlán"},
		"21160": {"21", "Teotlalco"},
		"21161": {"21", "Tepanco de López"},
		"21162": {"21", "Tepango de Rodríguez"},
		"21163": {"21", "Tepatlaxco de Hidalgo"},
		"21164": {"21", "Tepeaca"},
		"21165": {"21", "Tepemaxalco"},
		"21166": {"21", "Tepeojuma"},
		"21167": {"21", "Tepetzintla"},
		"21168": {"21", "Tepexco"},
		"21169": {"21", "Tepexi de Rodríguez"},
		"21170": {"21", "Tepeyahualco"},
		"21171": {"21", "Tepeyahualco de Cuauhtémoc"},
		"21172": {"21", "Tetela de Ocampo"},
		"21173": {"21", "Teteles de Avila Castillo"},
		"21174": {"21", "Teziutlán"},
		"21175": {"21", "Tianguismanalco"},
		"21176": {"21", "Tilapa"},
		"21177": {"21", "Tlacotepec de Benito Juárez"},
		"21178": {"21", "Tlacuilotepec"},
		"21179": {"21", "Tlachichuca"},
		"21180": {"21", "Tlahuapan"},
		"21181": {"21", "Tlaltenango"},
		"21182": {"21", "Tlanepantla"},
		"21183": {"21", "Tlaola"},
		"21184": {"21", "Tlapacoya"},
		"21185": {"21", "Tlapanalá"},
		"21186": {"21", "Tlatlauquitepec"},
		"21187": {"21", "Tlaxco"},
		"21188": {"21", "Tochimilco"},
		"21189": {"21", "Tochtepec"},
		"21190": {"21", "Totoltepec de Guerrero"},
		"21191": {"21", "Tulcingo"},
		"21192": {"21", "Tuzamapan de Galeana"},
		"21193": {"21", "Tzicatlacoyan"},
		"21194": {"21", "Venustiano Carranza"},
		"21195": {"21", "Vicente Guerrero"},
		"21196": {"21", "Xayacatlán de Bravo"},
		"21197": {"21", "Xicotepec"},
		"21198": {"21", "Xicotlán"},
		"21199": {"21", "Xiutetelco"},
		"21200": {"21", "Xochiapulco"},
		"21201": {"21", "Xochiltepec"},
		"21202": {"21", "Xochitlán de Vicente Suárez"},
		"21203": {"21", "Xochitlán Todos Santos"},
		"21204": {"21", "Yaonáhuac"},
		"21205": {"21", "Yehualtepec"},
		"21206": {"21", "Zacapala"},
		"21207": {"21", "Zacapoaxtla"},
		"21208": {"21", "Zacatlán"},
		"21209": {"21", "Zapotitlán"},
		"21210": {"21", "Zapotitlán de Méndez"},
		"21211": {"21", "Zaragoza"},
		"21212": {"21", "Zautla"},
		"21213": {"21", "Zihuateutla"},
		"21214": {"21", "Zinacatepec"},
		"21215": {"21", "Zongozotla"},
		"21216": {"21", "Zoquiapan"},
		"21217": {"21", "Zoquitlán"},
		"22001": {"22", "Amealco de Bonfil"},
		"22002": {"22", "Pinal de Amoles"},
		"22003": {"22", "Arroyo Seco"},
		"22004": {"22", "Cadereyta de Montes"},
		"22005": {"22", "Colón"},
		"22006": {"22", "Corregidora"},
		"22007": {"22", "Ezequiel Montes"},
		"22008": {"22", "Huimilpan"},
		"22009": {"22", "Jalpan de Serra"},
		"22010": {"22", "Landa de Matamoros"},
		"22011": {"22", "El Marqués"},
		"22012": {"22", "Pedro Escobedo"},
		"22013": {"22", "Peñamiller"},
		"22014": {"22", "Querétaro"},
		"22015": {"22", "San Joaquín"},
		"22016": {"22", "San Juan del Río"},
		"22017": {"22", "Tequisquiapan"},
		"22018": {"22", "Tolimán"},
		"23001": {"23", "Cozumel"},
		"23002": {"23", "Felipe Carrillo Puerto"},
		"23003": {"23", "Isla Mujeres"},
		"23004": {"23", "Othón P. Blanco"},
		"23005": {"23", "Benito Juárez"},
		"23006": {"23", "José María Morelos"},
		"23007": {"23", "Lázaro Cárdenas"},
		"23008": {"23", "Solidaridad"},
		"23009": {"23", "Tulum"},
		"23010": {"23", "Bacalar"},
		"24001": {"24", "Ahualulco"},
		"24002": {"24", "Alaquines"},
		"24003": {"24", "Aquismón"},
		"24004": {"24", "Armadillo de los Infante"},
		"24005": {"24", "Cárdenas"},
		"24006": {"24", "Catorce"},
		"24007": {"24", "Cedral"},
		"24008": {"24", "Cerritos"},
		"24009": {"24", "Cerro de San Pedro"},
		"24010": {"24", "Ciudad del Maíz"},
		"24011": {"24", "Ciudad Fernández"},
		"24012": {"24", "Tancanhuitz"},
		"24013": {"24", "Ciudad Valles"},
		"24014": {"24", "Coxcatlán"},
		"24015": {"24", "Charcas"},
		"24016": {"24", "Ebano"},
		"24017": {"24", "Guadalcázar"},
		"24018": {"24", "Huehuetlán"},
		"24019": {"24", "Lagunillas"},
		"24020": {"24", "Matehuala"},
		"24021": {"24", "Mexquitic de Carmona"},
		"24022": {"24", "Moctezuma"},
		"24023": {"24", "Rayón"},
		"24024": {"24", "Rioverde"},
		"24025": {"24", "Salinas"},
		"24026": {"24", "San Antonio"},
		"24027": {"24", "San Ciro de Acosta"},
		"24028": {"24", "San Luis Potosí"},
		"24029": {"24", "San Martín Chalchicuautla"},
		"24030": {"24", "San Nicolás Tolentino"},
		"24031": {"24", "Santa Catarina"},
		"24032": {"24", "Santa María del Río"},
		"24033": {"24", "Santo Domingo"},
		"24034": {"24", "San Vicente Tancuayalab"},
		"24035": {"24", "Soledad de Graciano Sánchez"},
		"24036": {"24", "Tamasopo"},
		"24037": {"24", "Tamazunchale"},
		"24038": {"24", "Tampacán"},
		"24039": {"24", "Tampamolón Corona"},
		"24040": {"24", "Tamuín"},
		"24041": {"24", "Tanlajás"},
		"24042": {"24", "Tanquián de Escobedo"},
		"24043": {"24", "Tierra Nueva"},
		"24044": {"24", "Vanegas"},
		"24045": {"24", "Venado"},
		"24046": {"24", "Villa de Arriaga"},
		"24047": {"24", "Villa de Guadalupe"},
		"24048": {"24", "Villa de la Paz"},
		"24049": {"24", "Villa de Ramos"},
		"24050": {"24", "Villa de Reyes"},
		"24051": {"24", "Villa Hidalgo"},
		"24052": {"24", "Villa Juárez"},
		"24053": {"24", "Axtla de Terrazas"},
		"24054": {"24", "Xilitla"},
		"24055": {"24", "Zaragoza"},
		"24056": {"24", "Villa de Arista"},
		"24057": {"24", "Matlapa"},
		"24058": {"24", "El Naranjo"},
		"25001": {"25", "Ahome"},
		"25002": {"25", "Angostura"},
		"25003": {"25", "Badiraguato"},
		"25004": {"25", "Concordia"},
		"25005": {"25", "Cosalá"},
		"25006": {"25", "Culiacán"},
		"25007": {"25", "Choix"},
		"25008": {"25", "Elota"},
		"25009": {"25", "Escuinapa"},
		"25010": {"25", "El Fuerte"},
		"25011": {"25", "Guasave"},
		"25012": {"25", "Mazatlán"},
		"25013": {"25", "Mocorito"},
		"25014": {"25", "Rosario"},
		"25015": {"25", "Salvador Alvarado"},
		"25016": {"25", "San Ignacio"},
		"25017": {"25", "Sinaloa"},
		"25018": {"25", "Navolato"},
		"26001": {"26", "Aconchi"},
		"26002": {"26", "Agua Prieta"},
		"26003": {"26", "Alamos"},
		"26004": {"26", "Altar"},
		"26005": {"26", "Arivechi"},
		"26006": {"26", "Arizpe"},
		"26007": {"26", "Atil"},
		"26008": {"26", "Bacadéhuachi"},
		"26009": {"26", "Bacanora"},
		"26010": {"26", "Bacerac"},
		"26011": {"26", "Bacoachi"},
		"26012": {"26", "Bácum"},
		"26013": {"26", "Banámichi"},
		"26014": {"26", "Baviácora"},
		"26015": {"26", "Bavispe"},
		"26016": {"26", "Benjamín Hill"},
		"26017": {"26", "Caborca"},
		"26018": {"26", "Cajeme"},
		"26019": {"26", "Cananea"},
		"26020": {"26", "Carbó"},
		"26021": {"26", "La Colorada"},
		"26022": {"26", "Cucurpe"},
		"26023": {"26", "Cumpas"},
		"26024": {"26", "Divisaderos"},
		"26025": {"26", "Empalme"},
		"26026": {"26", "Etchojoa"},
		"26027": {"26", "Fronteras"},
		"26028": {"26", "Granados"},
		"26029": {"26", "Guaymas"},
		"26030": {"26", "Hermosillo"},
		"26031": {"26", "Huachinera"},
		"26032": {"26", "Huásabas"},
		"26033": {"26", "Huatabampo"},
		"26034": {"26", "Huépac"},
		"26035": {"26", "Imuris"},
		"26036": {"26", "Magdalena"},
		"26037": {"26", "Mazatán"},
		"26038": {"26", "Moctezuma"},
		"26039": {"26", "Naco"},
		"26040": {"26", "Nácori Chico"},
		"26041": {"26", "Nacozari de García"},
		"26042": {"26", "Navojoa"},
		"26043": {"26", "Nogales"},
		"26044": {"26", "Onavas"},
		"26045": {"26", "Opodepe"},
		"26046": {"26", "Oquitoa"},
		"26047": {"26", "Pitiquito"},
		"26048": {"26", "Puerto Peñasco"},
		"26049": {"26", "Quiriego"},
		"26050": {"26", "Rayón"},
		"26051": {"26", "Rosario"},
		"26052": {"26", "Sahuaripa"},
		"26053": {"26", "San Felipe de Jesús"},
		"26054": {"26", "San Javier"},
		"26055": {"26", "San Luis Río Colorado"},
		"26056": {"26", "San Miguel de Horcasitas"},
		"26057": {"26", "San Pedro de la Cueva"},
		"26058": {"26", "Santa Ana"},
		"26059": {"26", "Santa Cruz"},
		"26060": {"26", "Sáric"},
		"26061": {"26", "Soyopa"},
		"26062": {"26", "Suaqui Grande"},
		"26063": {"26", "Tepache"},
		"26064": {"26", "Trincheras"},
		"26065": {"26", "Tubutama"},
		"26066": {"26", "Ures"},
		"26067": {"26", "Villa Hidalgo"},
		"26068": {"26", "Villa Pesqueira"},
		"26069": {"26", "Yécora"},
		"26070": {"26", "General Plutarco Elías Calles"},
		"26071": {"26", "Benito Juárez"},
		"26072": {"26", "San Ignacio Río Muerto"},
		"27001": {"27", "Balancán"},
		"27002": {"27", "Cárdenas"},
		"27003": {"27", "Centla"},
		"27004": {"27", "Centro"},
		"27005": {"27", "Comalcalco"},
		"27006": {"27", "Cunduacán"},
		"27007": {"27", "Emiliano Zapata"},
		"27008": {"27", "Huimanguillo"},
		"27009": {"27", "Jalapa"},
		"27010": {"27", "Jalpa de Méndez"},
		"27011": {"27", "Jonuta"},
		"27012": {"27", "Macuspana"},
		"27013": {"27", "Nacajuca"},
		"27014": {"27", "Paraíso"},
		"27015": {"27", "Tacotalpa"},
		"27016": {"27", "Teapa"},
		"27017": {"27", "Tenosique"},
		"28001": {"28", "Abasolo"},
		"28002": {"28", "Aldama"},
		"28003": {"28", "Altamira"},
		"28004": {"28", "Antiguo Morelos"},
		"28005": {"28", "Burgos"},
		"28006": {"28", "Bustamante"},
		"28007": {"28", "Camargo"},
		"28008": {"28", "Casas"},
		"28009": {"28", "Ciudad Madero"},
		"28010": {"28", "Cruillas"},
		"28011": {"28", "Gómez Farías"},
		"28012": {"28", "González"},
		"28013": {"28", "Güémez"},
		"28014": {"28", "Guerrero"},
		"28015": {"28", "Gustavo Díaz Ordaz"},
		"28016": {"28", "Hidalgo"},
		"28017": {"28", "Jaumave"},
		"28018": {"28", "Jiménez"},
		"28019": {"28", "Llera"},
		"28020": {"28", "Mainero"},
		"28021": {"28", "El Mante"},
		"28022": {"28", "Matamoros"},
		"28023": {"28", "Méndez"},
		"28024": {"28", "Mier"},
		"28025": {"28", "Miguel Alemán"},
		"28026": {"28", "Miquihuana"},
		"28027": {"28", "Nuevo Laredo"},
		"28028": {"28", "Nuevo Morelos"},
		"28029": {"28", "Ocampo"},
		"28030": {"28", "Padilla"},
		"28031": {"28", "Palmillas"},
		"28032": {"28", "Reynosa"},
		"28033": {"28", "Río Bravo"},
		"28034": {"28", "San Carlos"},
		"28035": {"28", "San Fernando"},
		"28036": {"28", "San Nicolás"},
		"28037": {"28", "Soto la Marina"},
		"28038": {"28", "Tampico"},
		"28039": {"28", "Tula"},
		"28040": {"28", "Valle Hermoso"},
		"28041": {"28", "Victoria"},
		"28042": {"28", "Villagrán"},
		"28043": {"28", "Xicoténcatl"},
		"29001": {"29", "Amaxac de Guerrero"},
		"29002": {"29", "Apetatitlán de Antonio Carvajal"},
		"29003": {"29", "Atlangatepec"},
		"29004": {"29", "Atltzayanca"},
		"29005": {"29", "Apizaco"},
		"29006": {"29", "Calpulalpan"},
		"29007": {"29", "El Carmen Tequexquitla"},
		"29008": {"29", "Cuapiaxtla"},
		"29009": {"29", "Cuaxomulco"},
		"29010": {"29", "Chiautempan"},
		"29011": {"29", "Muñoz de Domingo Arenas"},
		"29012": {"29", "Españita"},
		"29013": {"29", "Huamantla"},
		"29014": {"29", "Hueyotlipan"},
		"29015": {"29", "Ixtacuixtla de Mariano Matamoros"},
		"29016": {"29", "Ixtenco"},
		"29017": {"29", "Mazatecochco de José María Morelos"},
		"29018": {"29", "Contla de Juan Cuamatzi"},
		"29019": {"29", "Tepetitla de Lardizábal"},
		"29020": {"29", "Sanctórum de Lázaro Cárdenas"},
		"29021": {"29", "Nanacamilpa de Mariano Arista"},
		"29022": {"29", "Acuamanala de Miguel Hidalgo"},
		"29023": {"29", "Natívitas"},
		"29024": {"29", "Panotla"},
		"29025": {"29", "San Pablo del Monte"},
		"29026": {"29", "Santa Cruz Tlaxcala"},
		"29027": {"29", "Tenancingo"},
		"29028": {"29", "Teolocholco"},
		"29029": {"29", "Tepeyanco"},
		"29030": {"29", "Terrenate"},
		"29031": {"29", "Tetla de la Solidaridad"},
		"29032": {"29", "Tetlatlahuca"},
		"29033": {"29", "Tlaxcala"},
		"29034": {"29", "Tlaxco"},
		"29035": {"29", "Tocatlán"},
		"29036": {"29", "Totolac"},
		"29037": {"29", "Ziltlaltépec de Trinidad Sánchez Santos"},
		"29038": {"29", "Tzompantepec"},
		"29039": {"29", "Xaloztoc"},
		"29040": {"29", "Xaltocan"},
		"29041": {"29", "Papalotla de Xicohténcatl"},
		"29042": {"29", "Xicohtzinco"},
		"29043": {"29", "Yauhquemehcan"},
		"29044": {"29", "Zacatelco"},
		"29045": {"29", "Benito Juárez"},
		"29046": {"29", "Emiliano Zapata"},
		"29047": {"29", "Lázaro Cárdenas"},
		"29048": {"29", "La Magdalena Tlaltelulco"},
		"29049": {"29", "San Damián Texóloc"},
		"29050": {"29", "San Francisco Tetlanohcan"},
		"29051": {"29", "San Jerónimo Zacualpan"},
		"29052": {"29", "San José Teacalco"},
		"29053": {"29", "San Juan Huactzinco"},
		"29054": {"29", "San Lorenzo Axocomanitla"},
		"29055": {"29", "San Lucas Tecopilco"},
		"29056": {"29", "Santa Ana Nopalucan"},
		"29057": {"29", "Santa Apolonia Teacalco"},
		"29058": {"29", "Santa Catarina Ayometla"},
		"29059": {"29", "Santa Cruz Quilehtla"},
		"29060": {"29", "Santa Isabel Xiloxoxtla"},
		"30001": {"30", "Acajete"},
		"30002": {"30", "Acatlán"},
		"30003": {"30", "Acayucan"},
		"30004": {"30", "Actopan"},
		"30005": {"30", "Acula"},
		"30006": {"30", "Acultzingo"},
		"30007": {"30", "Camarón de Tejeda"},
		"30008": {"30", "Alpatláhuac"},
		"30009": {"30", "Alto Lucero de Gutiérrez Barrios"},
		"30010": {"30", "Altotonga"},
		"30011": {"30", "Alvarado"},
		"30012": {"30", "Amatitlán"},
		"30013": {"30", "Naranjos Amatlán"},
		"30014": {"30", "Amatlán de los Reyes"},
		"30015": {"30", "Angel R. Cabada"},
		"30016": {"30", "La Antigua"},
		"30017": {"30", "Apazapan"},
		"30018": {"30", "Aquila"},
		"30019": {"30", "Astacinga"},
		"30020": {"30", "Atlahuilco"},
		"30021": {"30", "Atoyac"},
		"30022": {"30", "Atzacan"},
		"30023": {"30", "Atzalan"},
		"30024": {"30", "Tlaltetela"},
		"30025": {"30", "Ayahualulco"},
		"30026": {"30", "Banderilla"},
		"30027": {"30", "Benito Juárez"},
		"30028": {"30", "Boca del Río"},
		"30029": {"30", "Calcahualco"},
		"30030": {"30", "Camerino Z. Mendoza"},
		"30031": {"30", "Carrillo Puerto"},
		"30032": {"30", "Catemaco"},
		"30033": {"30", "Cazones de Herrera"},
		"30034": {"30", "Cerro Azul"},
		"30035": {"30", "Citlaltépetl"},
		"30036": {"30", "Coacoatzintla"},
		"30037": {"30", "Coahuitlán"},
		"30038": {"30", "Coatepec"},
		"30039": {"30", "Coatzacoalcos"},
		"30040": {"30", "Coatzintla"},
		"30041": {"30", "Coetzala"},
		"30042": {"30", "Colipa"},
		"30043": {"30", "Comapa"},
		"30044": {"30", "Córdoba"},
		"30045": {"30", "Cosamaloapan de Carpio"},
		"30046": {"30", "Cosautlán de Carvajal"},
		"30047": {"30", "Coscomatepec"},
		"30048": {"30", "Cosoleacaque"},
		"30049": {"30", "Cotaxtla"},
		"30050": {"30", "Coxquihui"},
		"30051": {"30", "Coyutla"},
		"30052": {"30", "Cuichapa"},
		"30053": {"30", "Cuitláhuac"},
		"30054": {"30", "Chacaltianguis"},
		"30055": {"30", "Chalma"},
		"30056": {"30", "Chiconamel"},
		"30057": {"30", "Chiconquiaco"},
		"30058": {"30", "Chicontepec"},
		"30059": {"30", "Chinameca"},
		"30060": {"30", "Chinampa de Gorostiza"},
		"30061": {"30", "Las Choapas"},
		"30062": {"30", "Chocamán"},
		"30063": {"30", "Chontla"},
		"30064": {"30", "Chumatlán"},
		"30065": {"30", "Emiliano Zapata"},
		"30066": {"30", "Espinal"},
		"30067": {"30", "Filomeno Mata"},
		"30068": {"30", "Fortín"},
		"30069": {"30", "Gutiérrez Zamora"},
		"30070": {"30", "Hidalgotitlán"},
		"30071": {"30", "Huatusco"},
		"30072": {"30", "Huayacocotla"},
		"30073": {"30", "Hueyapan de Ocampo"},
		"30074": {"30", "Huiloapan de Cuauhtémoc"},
		"30075": {"30", "Ignacio de la Llave"},
		"30076": {"30", "Ilamatlán"},
		"30077": {"30", "Isla"},
		"30078": {"30", "Ixcatepec"},
		"30079": {"30", "Ixhuacán de los Reyes"},
		"30080": {"30", "Ixhuatlán del Café"},
		"30081": {"30", "Ixhuatlancillo"},
		"30082": {"30", "Ixhuatlán del Sureste"},
		"30083": {"30", "Ixhuatlán de Madero"},
		"30084": {"30", "Ixmatlahuacan"},
		"30085": {"30", "Ixtaczoquitlán"},
		"30086": {"30", "Jalacingo"},
		"30087": {"30", "Xalapa"},
		"30088": {"30", "Jalcomulco"},
		"30089": {"30", "Jáltipan"},
		"30090": {"30", "Jamapa"},
		"30091": {"30", "Jesús Carranza"},
		"30092": {"30", "Xico"},
		"30093": {"30", "Jilotepec"},
		"30094": {"30", "Juan Rodríguez Clara"},
		"30095": {"30", "Juchique de Ferrer"},
		"30096": {"30", "Landero y Coss"},
		"30097": {"30", "Lerdo de Tejada"},
		"30098": {"30", "Magdalena"},
		"30099": {"30", "Maltrata"},
		"30100": {"30", "Manlio Fabio Altamirano"},
		"30101": {"30", "Mariano Escobedo"},
		"30102": {"30", "Martínez de la Torre"},
		"30103": {"30", "Mecatlán"},
		"30104": {"30", "Mecayapan"},
		"30105": {"30", "Medellín"},
		"30106": {"30", "Miahuatlán"},
		"30107": {"30", "Las Minas"},
		"30108": {"30", "Minatitlán"},
		"30109": {"30", "Misantla"},
		"30110": {"30", "Mixtla de Altamirano"},
		"30111": {"30", "Moloacán"},
		"30112": {"30", "Naolinco"},
		"30113": {"30", "Naranjal"},
		"30114": {"30", "Nautla"},
		"30115": {"30", "Nogales"},
		"30116": {"30", "Oluta"},
		"30117": {"30", "Omealca"},
		"30118": {"30", "Orizaba"},
		"30119": {"30", "Otatitlán"},
		"30120": {"30", "Oteapan"},
		"30121": {"30", "Ozuluama de Mascareñas"},
		"30122": {"30", "Pajapan"},
		"30123": {"30", "Pánuco"},
		"30124": {"30", "Papantla"},
		"30125": {"30", "Paso del Macho"},
		"30126": {"30", "Paso de Ovejas"},
		"30127": {"30", "La Perla"},
		"30128": {"30", "Perote"},
		"30129": {"30", "Platón Sánchez"},
		"30130": {"30", "Playa Vicente"},
		"30131": {"30", "Poza Rica de Hidalgo"},
		"30132": {"30", "Las Vigas de Ramírez"},
		"30133": {"30", "Pueblo Viejo"},
		"30134": {"30", "Puente Nacional"},
		"30135": {"30", "Rafael Delgado"},
		"30136": {"30", "Rafael Lucio"},
		"30137": {"30", "Los Reyes"},
		"30138": {"30", "Río Blanco"},
		"30139": {"30", "Saltabarranca"},
		"30140": {"30", "San Andrés Tenejapan"},
		"30141": {"30", "San Andrés Tuxtla"},
		"30142": {"30", "San Juan Evangelista"},
		"30143": {"30", "Santiago Tuxtla"},
		"30144": {"30", "Sayula de Alemán"},
		"30145": {"30", "Soconusco"},
		"30146": {"30", "Sochiapa"},
		"30147": {"30", "Soledad Atzompa"},
		"30148": {"30", "Soledad de Doblado"},
		"30149": {"30", "Soteapan"},
		"30150": {"30", "Tamalín"},
		"30151": {"30", "Tamiahua"},
		"30152": {"30", "Tampico Alto"},
		"30153": {"30", "Tancoco"},
		"30154": {"30", "Tantima"},
		"30155": {"30", "Tantoyuca"},
		"30156": {"30", "Tatatila"},
		"30157": {"30", "Castillo de Teayo"},
		"30158": {"30", "Tecolutla"},
		"30159": {"30", "Tehuipango"},
		"30160": {"30", "Álamo Temapache"},
		"30161": {"30", "Tempoal"},
		"30162": {"30", "Tenampa"},
		"30163": {"30", "Tenochtitlán"},
		"30164": {"30", "Teocelo"},
		"30165": {"30", "Tepatlaxco"},
		"30166": {"30", "Tepetlán"},
		"30167": {"30", "Tepetzintla"},
		"30168": {"30", "Tequila"},
		"30169": {"30", "José Azueta"},
		"30170": {"30", "Texcatepec"},
		"30171": {"30", "Texhuacán"},
		"30172": {"30", "Texistepec"},
		"30173": {"30", "Tezonapa"},
		"30174": {"30", "Tierra Blanca"},
		"30175": {"30", "Tihuatlán"},
		"30176": {"30", "Tlacojalpan"},
		"30177": {"30", "Tlacolulan"},
		"30178": {"30", "Tlacotalpan"},
		"30179": {"30", "Tlacotepec de Mejía"},
		"30180": {"30", "Tlachichilco"},
		"30181": {"30", "Tlalixcoyan"},
		"30182": {"30", "Tlalnelhuayocan"},
		"30183": {"30", "Tlapacoyan"},
		"30184": {"30", "Tlaquilpa"},
		"30185": {"30", "Tlilapan"},
		"30186": {"30", "Tomatlán"},
		"30187": {"30", "Tonayán"},
		"30188": {"30", "Totutla"},
		"30189": {"30", "Tuxpan"},
		"30190": {"30", "Tuxtilla"},
		"30191": {"30", "Ursulo Galván"},
		"30192": {"30", "Vega de Alatorre"},
		"30193": {"30", "Veracruz"},
		"30194": {"30", "Villa Aldama"},
		"30195": {"30", "Xoxocotla"},
		"30196": {"30", "Yanga"},
		"30197": {"30", "Yecuatla"},
		"30198": {"30", "Zacualpan"},
		"30199": {"30", "Zaragoza"},
		"30200": {"30", "Zentla"},
		"30201": {"30", "Zongolica"},
		"30202": {"30", "Zontecomatlán de López y Fuentes"},
		"30203": {"30", "Zozocolco de Hidalgo"},
		"30204": {"30", "Agua Dulce"},
		"30205": {"30", "El Higo"},
		"30206": {"30", "Nanchital de Lázaro Cárdenas del Río"},
		"30207": {"30", "Tres Valles"},
		"30208": {"30", "Carlos A. Carrillo"},
		"30209": {"30", "Tatahuicapan de Juárez"},
		"30210": {"30", "Uxpanapa"},
		"30211": {"30", "San Rafael"},
		"30212": {"30", "Santiago Sochiapan"},
		"31001": {"31", "Abalá"},
		"31002": {"31", "Acanceh"},
		"31003": {"31", "Akil"},
		"31004": {"31", "Baca"},
		"31005": {"31", "Bokobá"},
		"31006": {"31", "Buctzotz"},
		"31007": {"31", "Cacalchén"},
		"31008": {"31", "Calotmul"},
		"31009": {"31", "Cansahcab"},
		"31010": {"31", "Cantamayec"},
		"31011": {"31", "Celestún"},
		"31012": {"31", "Cenotillo"},
		"31013": {"31", "Conkal"},
		"31014": {"31", "Cuncunul"},
		"31015": {"31", "Cuzamá"},
		"31016": {"31", "Chacsinkín"},
		"31017": {"31", "Chankom"},
		"31018": {"31", "Chapab"},
		"31019": {"31", "Chemax"},
		"31020": {"31", "Chicxulub Pueblo"},
		"31021": {"31", "Chichimilá"},
		"31022": {"31", "Chikindzonot"},
		"31023": {"31", "Chocholá"},
		"31024": {"31", "Chumayel"},
		"31025": {"31", "Dzán"},
		"31026": {"31", "Dzemul"},
		"31027": {"31", "Dzidzantún"},
		"31028": {"31", "Dzilam de Bravo"},
		"31029": {"31", "Dzilam González"},
		"31030": {"31", "Dzitás"},
		"31031": {"31", "Dzoncauich"},
		"31032": {"31", "Espita"},
		"31033": {"31", "Halachó"},
		"31034": {"31", "Hocabá"},
		"31035": {"31", "Hoctún"},
		"31036": {"31", "Homún"},
		"31037": {"31", "Huhí"},
		"31038": {"31", "Hunucmá"},
		"31039": {"31", "Ixil"},
		"31040": {"31", "Izamal"},
		"31041": {"31", "Kanasín"},
		"31042": {"31", "Kantunil"},
		"31043": {"31", "Kaua"},
		"31044": {"31", "Kinchil"},
		"31045": {"31", "Kopomá"},
		"31046": {"31", "Mama"},
		"31047": {"31", "Maní"},
		"31048": {"31", "Maxcanú"},
		"31049": {"31", "Mayapán"},
		"31050": {"31", "Mérida"},
		"31051": {"31", "Mocochá"},
		"31052": {"31", "Motul"},
		"31053": {"31", "Muna"},
		"31054": {"31", "Muxupip"},
		"31055": {"31", "Opichén"},
		"31056": {"31", "Oxkutzcab"},
		"31057": {"31", "Panabá"},
		"31058": {"31", "Peto"},
		"31059": {"31", "Progreso"},
		"31060": {"31", "Quintana Roo"},
		"31061": {"31", "Río Lagartos"},
		"31062": {"31", "Sacalum"},
		"31063": {"31", "Samahil"},
		"31064": {"31", "Sanahcat"},
		"31065": {"31", "San Felipe"},
		"31066": {"31", "Santa Elena"},
		"31067": {"31", "Seyé"},
		"31068": {"31", "Sinanché"},
		"31069": {"31", "Sotuta"},
		"31070": {"31", "Sucilá"},
		"31071": {"31", "Sudzal"},
		"31072": {"31", "Suma"},
		"31073": {"31", "Tahdziú"},
		"31074": {"31", "Tahmek"},
		"31075": {"31", "Teabo"},
		"31076": {"31", "Tecoh"},
		"31077": {"31", "Tekal de Venegas"},
		"31078": {"31", "Tekantó"},
		"31079": {"31", "Tekax"},
		"31080": {"31", "Tekit"},
		"31081": {"31", "Tekom"},
		"31082": {"31", "Telchac Pueblo"},
		"31083": {"31", "Telchac Puerto"},
		"31084": {"31", "Temax"},
		"31085": {"31", "Temozón"},
		"31086": {"31", "Tepakán"},
		"31087": {"31", "Tetiz"},
		"31088": {"31", "Teya"},
		"31089": {"31", "Ticul"},
		"31090": {"31", "Timucuy"},
		"31091": {"31", "Tinum"},
		"31092": {"31", "Tixcacalcupul"},
		"31093": {"31", "Tixkokob"},
		"31094": {"31", "Tixmehuac"},
		"31095": {"31", "Tixpéhual"},
		"31096": {"31", "Tizimín"},
		"31097": {"31", "Tunkás"},
		"31098": {"31", "Tzucacab"},
		"31099": {"31", "Uayma"},
		"31100": {"31", "Ucú"},
		"31101": {"31", "Umán"},
		"31102": {"31", "Valladolid"},
		"31103": {"31", "Xocchel"},
		"31104": {"31", "Yaxcabá"},
		"31105": {"31", "Yaxkukul"},
		"31106": {"31", "Yobaín"},
		"32001": {"32", "Apozol"},
		"32002": {"32", "Apulco"},
		"32003": {"32", "Atolinga"},
		"32004": {"32", "Benito Juárez"},
		"32005": {"32", "Calera"},
		"32006": {"32", "Cañitas de Felipe Pescador"},
		"32007": {"32", "Concepción del Oro"},
		"32008": {"32", "Cuauhtémoc"},
		"32009": {"32", "Chalchihuites"},
		"32010": {"32", "Fresnillo"},
		"32011": {"32", "Trinidad García de la Cadena"},
		"32012": {"32", "Genaro Codina"},
		"32013": {"32", "General Enrique Estrada"},
		"32014": {"32", "General Francisco R. Murguía"},
		"32015": {"32", "El Plateado de Joaquín Amaro"},
		"32016": {"32", "General Pánfilo Natera"},
		"32017": {"32", "Guadalupe"},
		"32018": {"32", "Huanusco"},
		"32019": {"32", "Jalpa"},
		"32020": {"32", "Jerez"},
		"32021": {"32", "Jiménez del Teul"},
		"32022": {"32", "Juan Aldama"},
		"32023": {"32", "Juchipila"},
		"32024": {"32", "Loreto"},
		"32025": {"32", "Luis Moya"},
		"32026": {"32", "Mazapil"},
		"32027": {"32", "Melchor Ocampo"},
		"32028": {"32", "Mezquital del Oro"},
		"32029": {"32", "Miguel Auza"},
		"32030": {"32", "Momax"},
		"32031": {"32", "Monte Escobedo"},
		"32032": {"32", "Morelos"},
		"32033": {"32", "Moyahua de Estrada"},
		"32034": {"32", "Nochistlán de Mejía"},
		"32035": {"32", "Noria de Ángeles"},
		"32036": {"32", "Ojocaliente"},
		"32037": {"32", "Pánuco"},
		"32038": {"32", "Pinos"},
		"32039": {"32", "Río Grande"},
		"32040": {"32", "Sain Alto"},
		"32041": {"32", "El Salvador"},
		"32042": {"32", "Sombrerete"},
		"32043": {"32", "Susticacán"},
		"32044": {"32", "Tabasco"},
		"32045": {"32", "Tepechitlán"},
		"32046": {"32", "Tepetongo"},
		"32047": {"32", "Teúl de González Ortega"},
		"32048": {"32", "Tlaltenango de Sánchez Román"},
		"32049": {"32", "Valparaíso"},
		"32050": {"32", "Vetagrande"},
		"32051": {"32", "Villa de Cos"},
		"32052": {"32", "Villa García"},
		"32053": {"32", "Villa González Ortega"},
		"32054": {"32", "Villa Hidalgo"},
		"32055": {"32", "Villanueva"},
		"32056": {"32", "Zacatecas"},
		"32057": {"32", "Trancoso"},
		"32058": {"32", "Santa María de la Paz"},
	}
)

func init() {
	log.SetFlags(0)
}

type State struct {
	Name          string  `json:"name"`
	PositiveCases int     `json:"positive"`
	NegativeCases int     `json:"negative"`
	SuspectCases  int     `json:"suspect"`
	Deaths        int     `json:"deaths"`
	AttackRate    float64 `json:"attack_rate"`
}

type Municipio struct {
	Name          string  `json:"name"`
	PositiveCases int     `json:"positive"`
	NegativeCases int     `json:"negative"`
	SuspectCases  int     `json:"suspect"`
	Deaths        int     `json:"deaths"`
	AttackRate    float64 `json:"attack_rate"`
}

type MunicipioDetail struct {
	EstadoGeo string
	Name      string
}

type SinaveData struct {
	States []State `json:"states"`
	tpc    int
	tnc    int
	tsc    int
	td     int

	// ar is the attackRate
	ar float64
}

func (s *SinaveData) UnmarshalJSON(b []byte) error {
	// More JSON is embedded into the object...
	// {"d":"[[]]"}
	var all map[string]interface{}
	err := json.Unmarshal(b, &all)
	if err != nil {
		return err
	}
	data := all["d"].(string)
	var states [][]interface{}
	err = json.Unmarshal([]byte(data), &states)
	if err != nil {
		return err
	}

	s.States = make([]State, 0)
	for _, entry := range states {
		// e.g.
		// [1 Aguascalientes 1353758.409 01 24 243 74 0]
		name := entry[1].(string)
		if name == "NACIONAL" {
			continue
		}

		pos, err := strconv.Atoi(entry[4].(string))
		if err != nil {
			return err
		}
		neg, err := strconv.Atoi(entry[5].(string))
		if err != nil {
			return err
		}
		susp, err := strconv.Atoi(entry[6].(string))
		if err != nil {
			return err
		}
		deaths, err := strconv.Atoi(entry[7].(string))
		if err != nil {
			return err
		}
		attackRate, err := strconv.ParseFloat(entry[8].(string), 64)
		if err != nil {
			return err
		}
		state := State{
			Name:          name,
			PositiveCases: pos,
			NegativeCases: neg,
			SuspectCases:  susp,
			Deaths:        deaths,
			AttackRate:    attackRate,
		}
		s.States = append(s.States, state)
	}

	return nil
}

func (sdata *SinaveData) TotalPositiveCases() int {
	if sdata.tpc > 0 {
		return sdata.tpc
	}
	for _, state := range sdata.States {
		sdata.tpc += state.PositiveCases
	}
	return sdata.tpc
}

func (sdata *SinaveData) TotalNegativeCases() int {
	if sdata.tnc > 0 {
		return sdata.tnc
	}
	for _, state := range sdata.States {
		sdata.tnc += state.NegativeCases
	}
	return sdata.tnc
}

func (sdata *SinaveData) TotalSuspectCases() int {
	if sdata.tsc > 0 {
		return sdata.tsc
	}
	for _, state := range sdata.States {
		sdata.tsc += state.SuspectCases
	}
	return sdata.tsc
}

func (sdata *SinaveData) TotalDeaths() int {
	if sdata.td > 0 {
		return sdata.td
	}
	for _, state := range sdata.States {
		sdata.td += state.Deaths
	}
	return sdata.td
}

func (sdata *SinaveData) TestPositivityRate() float64 {
	return float64(sdata.TotalPositiveCases()) / float64((sdata.TotalPositiveCases() + sdata.TotalNegativeCases()))
}

func fetchData(endpoint string) (*SinaveData, error) {
	hc := &http.Client{}
	req, err := http.NewRequest("POST", endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")

	resp, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error: %s", body)
	}

	var sdata *SinaveData
	err = json.Unmarshal(body, &sdata)
	if err != nil {
		return nil, err
	}
	return sdata, nil
}

func fetchPastData(endpoint string) (*SinaveData, error) {
	hc := &http.Client{}
	resp, err := hc.Get(endpoint)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error: %s", body)
	}

	type s struct {
		States []State `json:"states"`
	}
	var sd *s
	err = json.Unmarshal(body, &sd)
	if err != nil {
		log.Fatal(err)
	}

	sdata := &SinaveData{
		States: sd.States,
	}
	return sdata, nil
}

func detectLatestDataSource() (string, error) {
	hc := &http.Client{}
	resp, err := hc.Get(sinaveURL)
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Error: %s", body)
	}

	// ...
	if bytes.Contains(body, []byte("Grafica22")) {
		return sinaveURLA, nil
	}
	if bytes.Contains(body, []byte("Grafica23")) {
		return sinaveURLB, nil
	}
	return "", ErrSourceNotFound
}

func fetchMunicipalData(endpoint string, caseType string) (map[string]int, error) {
	vals := url.Values{"sPatType": {caseType}}
	resp, err := http.PostForm(endpoint, vals)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error: %s", body)
	}
	fmt.Println(string(body))
	muns, err := parseScript(string(body))
	if err != nil {
		return nil, err
	}

	return muns, nil
}

func parseScript(sample string) (map[string]int, error) {
	muns := map[string]int{}
	var (
		start, end   int
		vstart, vend int
		mun          string
	)

	reset := func(){
		start = 0
		end = 0
		vstart = 0
		vend = 0
	}
	for i, c := range sample {
		if start == 0 && c == 39 {
			start = i + 1
		} else if start > 0 && end == 0 && c == 39 {
			end = i
			mun = sample[start:end]
			if mun == "body" {
				break
			}
		} else if start > 0 && end > 0 && vstart == 0 && c == '=' {
			vstart = i + 1
		} else if start > 0 && end > 0 && vstart > 0 && vend == 0 && c == ';' {
			vend = i
			s := sample[vstart:vend]
			if s == ` new Array()` {
				reset()
				continue
			}
			
			v, err := strconv.Atoi(sample[vstart:vend])
			if err != nil {
				continue
				return nil, err
			}
			muns[mun] = v

			// Reset everything
			reset()
		}
	}

	return muns, nil
}

func showTable(sdata *SinaveData) {
	fmt.Println("|----------------------|-----------------|-----------------|-------------------|---------|-------------|------------|")
	fmt.Println("| Estado               | Casos Positivos | Casos Negativos | Casos Sospechosos | Decesos | Positividad | Incidencia |")
	fmt.Println("|----------------------|-----------------|-----------------|-------------------|---------|-------------|------------|")
	var totalAttackRate float64
	for _, state := range sdata.States {
		if state.Name == "NACIONAL" {
			totalAttackRate = state.AttackRate
			continue
		}
		testPositivityRate := float64(state.PositiveCases) / (float64(state.PositiveCases) + float64(state.NegativeCases))
		fmt.Printf("| %-20s | %-15d | %-15d | %-17d | %-7d | %-8.4f    | %-8.2f   |\n",
			state.Name,
			state.PositiveCases,
			state.NegativeCases,
			state.SuspectCases,
			state.Deaths,
			testPositivityRate,
			state.AttackRate,
		)
	}
	fmt.Println("|----------------------|-----------------|-----------------|-------------------|---------|-------------|------------|")
	fmt.Printf("| %-20s | %-15d | %-15d | %-17d | %-7d | %-8.4f    | %-8.4f   |\n",
		"TOTAL",
		sdata.TotalPositiveCases(),
		sdata.TotalNegativeCases(),
		sdata.TotalSuspectCases(),
		sdata.TotalDeaths(),
		sdata.TestPositivityRate(),
		totalAttackRate,
	)
	fmt.Println("|----------------------|-----------------|-----------------|-------------------|---------|-------------|------------|")
}

func showTableDiff(sdata, pdata *SinaveData) {
	pmap := make(map[string]State)
	for _, state := range pdata.States {
		if state.Name == "NACIONAL" {
			continue
		}
		pmap[state.Name] = state
	}

	fmt.Println("|----------------------|-----------------|-----------------|-------------------|-------------|")
	fmt.Println("| Estado               | Casos Positivos | Casos Negativos | Casos Sospechosos | Decesos     |")
	fmt.Println("|----------------------|-----------------|-----------------|-------------------|-------------|")
	for _, state := range sdata.States {
		pstate := pmap[state.Name]
		fmt.Printf("| %-20s | %-15s | %-15s | %-17s | %-11s |\n",
			state.Name,
			fmt.Sprintf("%-5d (%d)", state.PositiveCases-pstate.PositiveCases, state.PositiveCases),
			fmt.Sprintf("%-5d (%d)", state.NegativeCases-pstate.NegativeCases, state.NegativeCases),
			fmt.Sprintf("%-5d (%d)", state.SuspectCases-pstate.SuspectCases, state.SuspectCases),
			fmt.Sprintf("%-5d (%d)", state.Deaths-pstate.Deaths, state.Deaths),
		)
	}
	fmt.Println("|----------------------|-----------------|-----------------|-------------------|-------------|")
	fmt.Printf("| %-20s | %-15d | %-15d | %-17d | %-11d |\n",
		"TOTAL",
		sdata.TotalPositiveCases()-pdata.TotalPositiveCases(),
		sdata.TotalNegativeCases()-pdata.TotalNegativeCases(),
		sdata.TotalSuspectCases()-pdata.TotalNegativeCases(),
		sdata.TotalDeaths()-pdata.TotalDeaths(),
	)
	fmt.Println("|----------------------|-----------------|-----------------|-------------------|-------------|")
}

func showTableAwkFriendly(sdata *SinaveData) {
	for _, state := range sdata.States {
		if state.Name == "NACIONAL" {
			continue
		}
		var name string
		if state.Name == "Ciudad de México" {
			name = "CDMX"
		} else {
			name = strings.Join(strings.Fields(state.Name), "")
		}
		fmt.Printf("%-20s\t%-15d\t%-15d\t%-17d\t%-7d\n",
			name, state.PositiveCases, state.NegativeCases, state.SuspectCases, state.Deaths)
	}
}

func showJSON(sdata *SinaveData) {
	result, err := json.MarshalIndent(sdata, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(result))
}

func showCSV(sdata *SinaveData) {
	fmt.Println("\"Estado\"               , \"Casos Positivos\" , \"Casos Negativos\" , \"Casos Sospechosos\" , \"Decesos\"")
	for _, state := range sdata.States {
		if state.Name == "NACIONAL" {
			continue
		}
		fmt.Printf("  %-20s , %-15d , %-15d , %-17d , %-7d \n",
			state.Name, state.PositiveCases, state.NegativeCases, state.SuspectCases, state.Deaths)
	}
}

func showMunicipalData(config *CliConfig) error {
	state := config.municipio

	// Try to fetch by municipal data instead.
	pCases, err := fetchMunicipalData(municipalURL, "Confirmados")
	if err != nil {
		return err
	}
	nCases, err := fetchMunicipalData(municipalURL, "Negativos")
	if err != nil {
		return err
	}
	sCases, err := fetchMunicipalData(municipalURL, "Sospechosos")
	if err != nil {
		return err
	}
	dCases, err := fetchMunicipalData(municipalURL, "Defunciones")
	if err != nil {
		return err
	}
	states := make(map[string]State)
	muns := make(map[string]Municipio)

	// Collect positive, negative, suspect...
	for k, v := range pCases {
		muns[k] = Municipio{
			PositiveCases: v,
		}
	}
	for k, v := range nCases {
		m := muns[k]
		m.NegativeCases = v
		muns[k] = m
	}
	for k, v := range sCases {
		m := muns[k]
		m.SuspectCases = v
		muns[k] = m
	}
	for k, v := range dCases {
		m := muns[k]
		m.Deaths = v
		muns[k] = m
	}

	var tpCases, tnCases, tsCases, tdCases int

	if config.exportFormat != "json" {
		fmt.Println("|-------------------|-----------------|-----------------|-------------------|---------|-------------|---------------------------|")
		fmt.Println("| Estado            | Casos Positivos | Casos Negativos | Casos Sospechosos | Decesos | Positividad | Nombre                    |")
		fmt.Println("|-------------------|-----------------|-----------------|-------------------|---------|-------------|---------------------------|")
	}
	for s, m := range muns {
		details := MunicipiosMexico[s]
		filter := s[:2]
		sName := StatesMap[filter]

		stateName := strings.Join(strings.Fields(sName), "")
		if s, ok := states[filter]; ok {
			s.PositiveCases += m.PositiveCases
			s.NegativeCases += m.NegativeCases
			s.SuspectCases += m.SuspectCases
			s.Deaths += m.Deaths
			states[filter] = s
		} else {
			states[filter] = State{
				Name:          StatesMap[filter],
				PositiveCases: m.PositiveCases,
				NegativeCases: m.NegativeCases,
				SuspectCases:  m.SuspectCases,
				Deaths:        m.Deaths,
			}
		}

		// Skip match all filters but allow narrow down per state.
		if state != "*" && state != "all" && state != filter {
			continue
		}

		var positivity float64
		if m.PositiveCases > 0 {
			positivity = float64(m.PositiveCases) / float64(m.PositiveCases+m.NegativeCases)
		}
		tpCases += m.PositiveCases
		tnCases += m.NegativeCases
		tsCases += m.SuspectCases
		tdCases += m.Deaths
		if config.exportFormat != "json" {
			fmt.Printf("| %-17s | %-15d | %-15d | %-17d | %-7d | %-11.4f | %s\n",
				stateName, m.PositiveCases, m.NegativeCases, m.SuspectCases, m.Deaths, positivity, details.Name)
		}
	}
	totalPositivity := float64(tpCases) / float64(tpCases+tnCases)

	if config.exportFormat != "json" {
		fmt.Println("|-------------------|-----------------|-----------------|-------------------|---------|-------------|")
		fmt.Printf("| %-17s | %-15d | %-15d | %-17d | %-7d | %-11.4f |\n",
			"TOTAL", tpCases, tnCases, tsCases, tdCases, totalPositivity)
		fmt.Println("|-------------------|-----------------|-----------------|-------------------|---------|-------------|")
	}
	sdata := &SinaveData{
		States: make([]State, 0),
	}
	for _, v := range states {
		sdata.States = append(sdata.States, v)
	}
	if config.municipio == "states" {
		sdata2, err := fetchData(attackRateURL)
		if err != nil {
			return err
		}
		for i, s := range sdata.States {
			for _, s2 := range sdata2.States {
				if s2.Name == s.Name {
					s.AttackRate = s2.AttackRate
					sdata.States[i] = s
				}

			}
		}

		switch config.exportFormat {
		case "json":
			showJSON(sdata)
		case "table":
			showTable(sdata)
		}
	}
	return nil
}

type CliConfig struct {
	showVersion  bool
	showHelp     bool
	exportFormat string
	source       string
	since        string
	municipio    string
}

func main() {
	fs := flag.NewFlagSet("covid19mx", flag.ExitOnError)
	flag.Usage = func() {
		fmt.Printf("Usage: covid19mx [options...]\n\n")
		fs.PrintDefaults()
		fmt.Println()
	}

	// Top level global config
	config := &CliConfig{}
	fs.BoolVar(&config.showHelp, "h", false, "Show help")
	fs.BoolVar(&config.showHelp, "help", false, "Show help")
	fs.BoolVar(&config.showVersion, "version", false, "Show version")
	fs.BoolVar(&config.showVersion, "v", false, "Show version")
	fs.StringVar(&config.exportFormat, "o", "", "Export format (options: json, csv, table)")
	fs.StringVar(&config.source, "source", "", "Source of the data")
	fs.StringVar(&config.since, "since", "", "Date against which to compare the data")
	fs.StringVar(&config.municipio, "municipio", "", "Municipio used to narrow down data")
	fs.StringVar(&config.municipio, "mun", "", "Municipio used to narrow down data")
	fs.Parse(os.Args[1:])

	switch {
	case config.showHelp:
		flag.Usage()
		os.Exit(0)
	case config.showVersion:
		fmt.Printf("covid19mx v%s\n", version)
		fmt.Printf("Release-Date %s\n", releaseDate)
		os.Exit(0)
	}

	if config.municipio != "" {
		err := showMunicipalData(config)
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}

	var (
		sdata *SinaveData
		err   error
	)
	if strings.Contains(config.source, ".json") {
		// Use a local file as the source
		data, err := ioutil.ReadFile(config.source)
		if err != nil {
			log.Fatal(err)
		}
		type s struct {
			States []State `json:"states"`
		}
		var sd *s
		err = json.Unmarshal(data, &sd)
		if err != nil {
			log.Fatal(err)
		}
		sdata = &SinaveData{
			States: sd.States,
		}
	} else {
		// Get latest sinave data by default.  Can also use a local checked
		// version for the data or an explicit http endpoint.
		if config.source == "" {
			config.source = attackRateURL
		}
		sdata, err = fetchData(config.source)
		if err != nil {
			log.Fatal(err)
		}
	}

	if config.since != "" {
		var date time.Time
		switch config.since {
		case "-1d", "1d", "yesterday":
			date = time.Now().AddDate(0, 0, -1)
		case "-2d", "2d", "2 days ago":
			date = time.Now().AddDate(0, 0, -2)
		default:
			days, err := strconv.Atoi(config.since)
			if err != nil {
				log.Fatal(err)
			}
			date = time.Now().AddDate(0, 0, days*-1)
		}
		pdata, err := fetchPastData(repoURL + date.Format("2006-01-02") + ".json")
		if err != nil {
			log.Fatal(err)
		}
		showTableDiff(sdata, pdata)
	} else {
		switch config.exportFormat {
		case "csv":
			showCSV(sdata)
		case "json":
			showJSON(sdata)
		case "table":
			showTable(sdata)
		case "awk":
			showTableAwkFriendly(sdata)
		default:
			showTable(sdata)
		}
	}
}
