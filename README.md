![hifi-baby](doc/hifi-baby.png)

Hifi Baby est une mini chaine hifi et un projet musical destinée aux enfants.

Le projet consiste à diffuser de la musique aléatoirement à la demande de l'enfant (appuie sur un bouton).

Les morceaux peuvent être ajoutés / supprimés via l'interface.

## Organisation du projet

Le programme est composé :
 - d'une application en GO (à la racine) contenant :
   - une API en HTTP REST/JSON
   - une interface GPIO (spécifique raspberry-pi)
   - une interface audio (via https://github.com/gopxl/beep)
 - d'une interface graphique en Vite/Vue/PrimeVue

## Application GO

Génération de l'application

```bash
go build
```

L'API est disponible au format postman : [doc/HifiBaby.postman_collection.json](doc/HifiBaby.postman_collection.json)


Variables d'environnements

| Module    | Variable          | Description                                | Valeur par défaut          |
|-----------|-------------------|--------------------------------------------|----------------------------|
| Général   | LOG_LEVEL         | Niveau de log                              | info                       |
| Général   | SETTINGS_PATH     | Chemin vers le fichier de configuration    | settings.json              |
| Audio     | STORAGE_PATH      | Chemin de stockage des pistes audio        | tracks                     |
| Serveur   | SERVER_URL        | URL du serveur                             | localhost:3000             |
| Serveur   | SERVER_UI_PATH    | Chemin vers l'interface utilisateur        | dist                       |
| Base de données | DATABASE_PATH | Chemin vers le fichier de la base de données | ./hifi-baby.db         |
| Base de données | DATABASE_TIMEOUT | Délai d'expiration pour la base de données | 10s                   |

### Requirements

```bash
# To enable databases : install sqlite3
sudo apt install sqlite3
```

