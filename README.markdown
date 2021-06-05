# Tankerkönig InfluxDB Importer

Exports Tankerkönig data to InfluxDB

# Build

```command
$ goreleaser --snapshot
```

# Deploy

```command
$ cd deployment
$ ansible-playbook playbook.yml -i somewhere.example.com,
```

# Manual Approach

1. Search for a station

    ```command
    $ curl https://creativecommons.tankerkoenig.de/json/list.php?lat=48.521&lng=8.82&rad=15&sort=dist&type=all&apikey=00000000-0000-0000-0000-000000000002
    ...
    870efffb-676b-4301-854e-c80e93c3e3ef
    ```

1. For that station, get the current price of Diesel:

    ```command
    $ TANKERKOENIG_STATIONS=870efffb-676b-4301-854e-c80e93c3e3ef
    $ curl "https://creativecommons.tankerkoenig.de/json/prices.php?ids=$TANKERKOENIG_STATIONS&apikey=$TANKERKOENIG_API_KEY" | jq -r '.prices[].diesel'
    ```

# Ideas

* Concourse-Resource produziert einen neuen Preis, wenn er sich verändert hat. Darauf kann z.B. eine Slack-Nachricht folgen.

# Links

* Data from [Tankerkönig](https://creativecommons.tankerkoenig.de/)
