# Go demo

By [Jesus Gomez](http://jesusgomez.io/).

Develop a command-line tool for Magic the Gathering. The source of data will be the API <https://api.magicthegathering.io/v1/cards> without using any query parameters for filtering.

## Use cases

-   Returns a list of **Cards** grouped by _set_

-   Returns a list of **Cards** grouped by _set_ and then each _set_ grouped by _rarity_

-   Returns a list of cards from the **Khans of Tarkir (KTK)** that ONLY have the colours _red_ AND _blue_ (same results as https://api.magicthegathering.io/v1/cards?set=KTK&colors=Red,Blue)

## How to use it

Docker is required to ensure no divergence between environments.

```sh
docker run -it --rm \
    -v "$(pwd):/usr/src/app" \
    -w /usr/src/app \
    golang:1.14-stretch \
    sh -c 'go run main.go [FLAG]'
```

Where FLAG can be:

-   `-help` or empty: show the CLI help
-   `-set` or `--set` Returns a list of Cards grouped by set
-   `-set-rarity` or `--set-rarity` Returns a list of Cards grouped by set and then each set grouped by rarity
-   `-ktk` or `--ktk` Returns a list of cards from the Khans of Tarkir (KTK) that ONLY have the colours red AND blue

To run the application tests:

```sh
docker run -it --rm \
    -v "$(pwd):/usr/src/app" \
    -w /usr/src/app \
    golang:1.14-stretch \
    sh -c 'go test -v ./...'
```

## Notes

-   [Official Go code conventions](https://github.com/golang/go/wiki/CodeReviewComments)

-   Only standard library

-   Parallelising to speed up fetching respecting the API's Rate Limiting
