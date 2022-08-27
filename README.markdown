# Tankerkönig Exporter

Exports Tankerkönig data for Prometheus

# Build

```command
$ goreleaser --snapshot
```

# Deployment

```command
$ cd deployment
$ ansible-playbook playbook.yml -i somewhere.example.com,
```

# Manual Approach

1. Search for a station within 1 km of my home:

    ```command
    $ curl "https://creativecommons.tankerkoenig.de/json/list.php?lat=48.52&lng=8.82&rad=1&sort=dist&type=all&apikey=$TANKERKOENIG_API_KEY" | jq -r '.stations[].id'
    870efffb-676b-4301-854e-c80e93c3e3ef
    ```

1. For that station, get the current price of Diesel:

    ```command
    $ TANKERKOENIG_STATIONS=870efffb-676b-4301-854e-c80e93c3e3ef
    $ curl "https://creativecommons.tankerkoenig.de/json/prices.php?ids=$TANKERKOENIG_STATIONS&apikey=$TANKERKOENIG_API_KEY" | jq -r '.prices[].diesel'
    ```

# Ideas

* Concourse resource creates a new version when the price has change (risen/fallen)

# Links

* Data from [Tankerkönig](https://creativecommons.tankerkoenig.de/)
