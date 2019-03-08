# KumuluzEE Go Samples

> These samples demonstrate how to get started using KumuluzEE microservice framework for Go microservices. They provide small, specific, working samples that can be used as a reference for your own projects.

These samples and quickstarts contain several working projects that demonstrate how to use [KumuluzEE](https://github.com/kumuluz/kumuluzee) Go packages. They also serve as test projects for the framework itself.

We recommend that you go through these samples to get a better understanding of the Go packages for the KumuluzEE framework and use them as a reference for your own projects.

Keep in mind that while projects containing multiple microservices are located in the same repository in order to simplify things, is is often recommended that you separate microservices by repository as well.

## About

The samples demonstrate simple use cases for using KumuluzEE Go packages to create self-sustaining microservices. The latest version of the samples will always use the latest version of the KumuluzEE framework. Therefore, it is recommended to use the latest version of the KumuluzEE framework for these samples. This way, you will also get all the latest features of the KumuluzEE. Refer to the usage section on how to build and run the samples.

Some samples are tagged as well. The tags (eg. `v2.2.0`) will correspond to the KumuluzEE release version in order to easily access the desired version of the framework that is used in the examples. The `master` branch will always use the latest snapshot version of the framework and the latest samples.

The following samples are available (list might not be up-to-date; please refer to the actual list above):

Go samples:
- KumuluzEE Config
- KumuluzEE Discovery
- Tutorial with Go and Java services

## Requirements

In order to run these examples as they are intended, you will need the following:

1. Go version >= 1.10.0 installed (suggested, this is version package is tested in)
    * If you have Go installed, you can check your version by typing the following in command line:
    ```
    $ go version
    ```
2. Git:
    * If you have installed Git, you can check the version by typing the following in a command line:
    ```
    $ git --version
    ```
        
## Usage

1. Clone the Git repository containing the examples:

    ```
    git clone git@https://github.com/kumuluz/kumuluzee-go-samples.git
    ```
    
2. Checkout the desired tagged version of the examples and the KumuluzEE framework (alternatively skip this step if you want the latest and greatest)

    ```
    cd kumuluzee-go-samples
    git checkout v2.4.0
    ```

To run a specific sample, please refer to the specific README file of the sample.
Most of the time you either run it directly with go toolchain or build Docker containers and run them.

## Changelog

Recent changes can be viewed on Github on the [Releases Page](https://github.com/kumuluz/kumuluzee-go-samples/releases)

## Contribute

See the [contributing docs](https://github.com/kumuluz/kumuluzee-go-samples/blob/master/CONTRIBUTING.md)

When submitting an issue, please follow the [guidelines](https://github.com/kumuluz/kumuluzee-go-samples/blob/master/CONTRIBUTING.md#bugs).

Issues related to KumuluzEE itself should be submitted at Issues page of appropriate library.

## License

MIT
