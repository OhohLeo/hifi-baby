![hifi-baby](doc/hifi-baby.png)

Hifi Baby est une mini chaine hifi et un projet musical destinée aux enfants.

Le projet consiste à diffuser de la musique aléatoirement à la demande de l'enfant (appuie sur un bouton).

Les morceaux peuvent être ajoutés / supprimés via l'interface.

## Organisation du projet

Le programme est composé :
 - d'une application en GO (à la racine) contenant :
   - une API en HTTP REST/JSON
   - un serveur Bluetooth Low Energy (BLE) GATT
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
| Bluetooth | BLE_ENABLED       | Activer/désactiver le serveur BLE          | true                       |
| Bluetooth | BLE_DEVICE_NAME   | Nom du périphérique Bluetooth              | HiFi Baby                  |
| Bluetooth | BLE_DISCOVERABLE  | Mode découvrable                           | true                       |
| Bluetooth | BLE_MTU           | Taille maximale des paquets (octets)       | 512                        |
| Bluetooth | BLE_CHUNK_SIZE    | Taille des chunks pour transfert (octets)  | 480                        |
| Bluetooth | BLE_STORAGE_PATH  | Chemin de stockage pour uploads BLE        | tracks                     |
| Base de données | DATABASE_PATH | Chemin vers le fichier de la base de données | ./hifi-baby.db         |
| Base de données | DATABASE_TIMEOUT | Délai d'expiration pour la base de données | 10s                   |

### Requirements

```bash
# To enable databases : install sqlite3
sudo apt install sqlite3
```

## Bluetooth Low Energy (BLE)

Le projet inclut un serveur Bluetooth GATT permettant de contrôler l'intégralité des fonctionnalités via BLE, offrant une alternative sans fil à l'API HTTP.

### Fonctionnalités

- **Contrôle audio** : Lecture, pause, arrêt, changement de position
- **Gestion du volume** : Augmentation, diminution, mode silencieux
- **Gestion des pistes** : Upload de fichiers audio (transfert par chunks), suppression
- **Statistiques** : Historique d'écoute, pistes les plus écoutées
- **Paramètres** : Lecture et modification des réglages
- **Notifications temps réel** : État du lecteur, progression des uploads

### Client Python

Un client complet est fourni dans `scripts/bluetooth_client.py` :

```bash
# Installation des dépendances
cd scripts
pip3 install -r requirements.txt

# Exemples d'utilisation
./bluetooth_client.py tracks              # Liste les pistes
./bluetooth_client.py state               # État du lecteur
./bluetooth_client.py play <track-id>     # Lecture d'une piste
./bluetooth_client.py pause               # Pause
./bluetooth_client.py volume-up           # Augmenter le volume
./bluetooth_client.py upload track.mp3    # Upload d'un fichier
./bluetooth_client.py listen              # Écoute des notifications
```

### Démarrage avec Bluetooth

**Méthode recommandée : utiliser le script de démarrage**
```bash
./scripts/start_bluetooth.sh
```

Ce script configure automatiquement l'adaptateur Bluetooth et lance l'application.

**Démarrage manuel :**
```bash
# 1. Configurer l'adaptateur (OBLIGATOIRE avant de lancer l'app)
bluetoothctl power on
bluetoothctl discoverable on
bluetoothctl pairable on

# 2. Ajouter les capacités
sudo setcap 'cap_net_raw,cap_net_admin+eip' ./hifi-baby

# 3. Lancer l'application
./hifi-baby
```

**Important :** L'adaptateur Bluetooth doit être en mode "découvrable" AVANT de lancer l'application, sinon la publicité BLE échouera.

### Documentation complète

Voir [bluetooth/README.md](bluetooth/README.md) pour :
- Architecture détaillée des services GATT
- Protocole de communication
- UUIDs des services et caractéristiques
- Format des messages JSON
- Protocole de transfert de fichiers
- Dépannage et tests

