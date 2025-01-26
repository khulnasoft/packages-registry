# Packages Importer

Packages Importer is an open-source CLI tool. You can use Packages Importer to copy Packages
between two or more Package registries.

This CLI is configured with a YAML file that describes:

* The source package registry.
* The destination package registry.
* The packages to copy from the source to the destination.

## Usage

### Locally

1. Download the executable from the [releases page](https://github.com/khulnasoft/packages-registry/-/releases).
1. Create a `config.yml` file defining a source and a destination for a single package:

  ```yaml
  my_example:
    type: npm
    source:
      url: http://source.registry.example/npm
      credentials:
        token: $SOURCE_TOKEN
    destination:
      url: http://destination.registry.example/npm
      credentials:
        token: $DESTINATION_TOKEN
    packages:
      "@my_company/my_package": 4.2.7
  ```

1. Run `pkgs_importer generate`:

  ```shell
  pkgs_importer generate
  ```

1. Packages Bus generates a KhulnaSoft pipeline configuration. You can use `less` to view the configuration file:

  ```shell
  less child_pipeline.yml
  ```

This configuration file generates a pipeline configuration (`child_pipeline.yml`)
that defines a single job (copy job) that copies the NPM package
`@my_company/my_package`, version `4.2.7` from
`http://source.registry.example/npm` to `http://destination.registry.example/npm`.

You can define as many sources, destinations, and packages as you need. However,
you must stay within the [limits](#cicd-limitations) of the KhulnaSoft pipelines.


A configuration file contains import blocks similar to this example above. A import block consists of the following:

- A package format. For more information about the values you can use, see the [formats supported](#formats-supported).
- A source packages registry. You must define a `<source_url>`. Credentials are optional.
- A destination packages registry. You must define a `<destination_url>`. Credentials are required.
- A [set of packages](#describing-packages).

### Describing packages

You can describe packages in 3 forms.

You can use tuples of package names and versions. For example:

```yaml
packages:
  package_1: 1.2.3
  package_2: 2.5.6
  package_3: 4.5.8
  package_3: 7.8.1
  package_3: 10.5.9
```

You can pack versions of the same package together. For example:

```yaml
packages:
  package_1: 1.2.3
  package_2: 2.5.6
  package_3:
    - 4.5.8
    - 7.8.1
    - 10.5.9
```

You can point to a `.csv` file that describes your packages. For example:

```yaml
packages: "packages.csv"
```

The following is the `.csv` file of the previous example:

```csv
package_1,1.2.3
package_2,2.5.6
package_3,4.5.8
package_3,7.8.1
package_3,10.5.9
```

NOTE:
There are no headings or titles in the `.csv` file.

## Formats supported

* [NPM](#npm)
* [NuGet](#nuget)
* [Maven](#maven)
* [PyPI](#pypi)

### NPM

Prerequisite:

- Your version of `npm` is at least 7.24.2.

The default image used for importing jobs is: [`node:latest`](https://hub.docker.com/_/node).

You can add optional credential fields to the `.npmrc` file. For more information, see the
available [configuration options](https://docs.npmjs.com/cli/v6/using-npm/config#config-settings) for NPM.

The following examples show `source` package registries, but you can also use them as
`destination` registries.

#### Limitations

If a `publishConfig` is set with a registry url, `$ npm publish` will not publish the package elsewhere.
See this [section](https://docs.npmjs.com/cli/v9/configuring-npm/package-json#publishconfig) and this [section](https://docs.npmjs.com/cli/v9/using-npm/registry#how-can-i-prevent-my-package-from-being-published-in-the-official-registry) of the NPM documentation.

To workaround this issue, this CLI app will open the package and remove the `publishConfig`.

This workaround can result in a change of the checksums reported by `$ npm`.

#### KhulnaSoft

```yaml
source:
  url: https://khulnasoft.example.com/api/v4/projects/<project_id>/packages/npm/
  credentials:
    token: $TOKEN
```

`$TOKEN` can be one of the following [tokens](https://docs.khulnasoft.com/ee/user/packages/package_registry/supported_functionality.html#authentication-tokens),
saved as [an environment variable](#use-environment-variables-for-your-token-values):

- A personal token.
- A deploy token.
- A CI job token.

WARNING:
The URL must end with `/`. The package copy fails if the trailing slash is missing.

[Duplicates are not allowed](https://docs.khulnasoft.com/ee/user/packages/npm_registry/#package-already-exists) with NPM.

If the tool tries to publish the package to a KhulnaSoft project where it already
exists, the publication will be rejected.

#### Github

```yaml
source:
  url: https://npm.pkg.github.com
  credentials:
    token: $TOKEN
```

For more information on which `$TOKEN` can be used, see the [GitHub documentation](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-npm-registry#authenticating-to-github-packages).

#### Artifactory

```yaml
source:
  url: https://<namespace>.jfrog.io/artifactory/api/npm/default-npm/
  credentials:
    token: $TOKEN
    _base64_token: true
    email: <user_email>
    always-auth: true
```

For more information on which `$TOKEN` can be used, see the [JFrog documentation](https://www.jfrog.com/confluence/display/RTF/Npm+Registry#npmRegistry-UsingBasicAuthentication).

It is important to note that:

- Artifactory registries use base64 tokens. You must set your `_base64_token` to `true` so that
this tool will send it properly.
- Artifactory registries require `GET` requests to be authenticated. You must set `always-auth`
to `true` and set `email` to your user email.

### NuGet

The default image used for importing jobs is [`mono:6`](https://hub.docker.com/_/mono).

In the credentials configuration, any additional fields are passed to the `nuget sources` command used to set up authentication.
See the [`nuget sources` reference](https://learn.microsoft.com/en-us/nuget/reference/cli-reference/cli-ref-sources) for more information.

The following examples show `source` package registries, but you can also use them as
`destination` registries.

#### NuGet Limitations

[NuGet symbol packages](https://learn.microsoft.com/en-us/nuget/create-packages/symbol-packages-snupkg) are not supported.

The configuration file package names are automatically transformed to lower case due to a [technical constraint](https://github.com/spf13/viper#does-viper-support-case-sensitive-keys).
Importing NuGet packages works fine with camelCase or lowercase package names. If this is an issue for your use case, a possible workaround is to use the [`.csv`](#describing-packages) file to list packages to import.

Using the `$ nuget setapikey` command with API keys are not supported. Only username and password combinations are supported.

#### KhulnaSoft

```yaml
source:
  url: https://khulnasoft.example.com/api/v4/projects/<your_project_id>/packages/nuget/index.json
  credentials:
    username: $USERNAME
    token: $TOKEN
```

`$TOKEN` can be one of the following [tokens](https://docs.khulnasoft.com/ee/user/packages/package_registry/supported_functionality.html#authentication-tokens),
saved as [an environment variable](#use-environment-variables-for-your-token-values):

- A personal access token.
- A deploy token.
- A CI/CD job token.

`$USERNAME` is required and is the username that `nuget sources` uses to set up authentication.

#### Github

```yaml
source:
  url: https://nuget.pkg.github.com/<namespace>/index.json
  credentials:
    username: $USERNAME
    token: $TOKEN
```

For more information on which `$TOKEN` can be used, see the [GitHub documentation](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-nuget-registry#authenticating-to-github-packages).

#### Artifactory

```yaml
source:
  url: https://<namespace>.jfrog.io/artifactory/api/npm/default-npm/
  credentials:
    username: $USERNAME
    token: $TOKEN
```

To get, the `$USERNAME` and `$TOKEN` values, use the [set me up instructions](https://jfrog.com/knowledge-base/the-set-me-up-option-explained/) from your Artifactory NuGet repository.

Do not try use the `nuget setapikey` instructions that are documented in the [NuGet registry authentication](https://www.jfrog.com/confluence/display/JFROG/NuGet+Repositories#NuGetRepositories-NuGetAuthentication).
This tool doesn't support api keys, see the [known limitations](#nuget-limitations).

### Maven

The default image used for importing jobs is [`maven:eclipse-temurin`](https://hub.docker.com/_/maven).

In the credentials configuration, you can either:

- Pass a `token` and a `username`. They are used to generate [`<username>` and `<password>` sections](https://maven.apache.org/settings.html#servers) in the `settings.xml` file.
- Pass a `token` and a `header_name`. They are used to generate a [`<httpHeaders>` `<property>` section](https://maven.apache.org/guides/mini/guide-http-settings.html#http-headers) in the `settings.xml` file.

See the examples below to know which configuration to use with each Maven package registry provider.

The following examples show `source` package registries, but you can also use them as
`destination` registries.

#### Describing packages

Maven packages are identified by their [Maven coordinates](https://maven.apache.org/pom.html#Maven_Coordinates) which are essentially:

- The group ID.
- The artifact ID.
- The version.
- Optional. The packaging.

To define the packages to import with this tool, join the group ID and the artifact ID with a `:`, and quote the result.

The packaging must be one of the following:

- `pom`
- `jar`
- `maven-plugin`
- `ejb`
- `war`
- `ear`
- `rar`
- `aar`

By default, the `jar` packaging is used.

To specify a packaging, join the version and the packaging with `:`, and quote the result.

For example, to import package group ID `com.my.company`, artifact ID `my.fine.package`, and version `1.2.3`,
you must use:

```yaml
packages:
  "com.my.company:my.fine.package": 1.2.3
  "com.my.company:my.second.package": "2.6.8:pom" # pom only package
  "com.my.company:my.app": "7.4.9:war"
```

Similarly, if you use a CSV to describe packages, you need to use:

```csv
com.my.company:my.fine.package,1.2.3
```

#### Maven Limitations

Artifacts built with [classifiers](https://maven.apache.org/pom.html#dependencies) like `-javadoc.jar` or `-sources.jar` are not supported.

#### KhulnaSoft

```yaml
source:
  url: https://khulnasoft.example.com/api/v4/projects/<your_project_id>/packages/maven
  credentials:
    token: $TOKEN
    header_name: <header-name>
```

`$TOKEN` can be one of the following [tokens](https://docs.khulnasoft.com/ee/user/packages/package_registry/supported_functionality.html#authentication-tokens),
saved as [an environment variable](#use-environment-variables-for-your-token-values):

- A personal access token.
- A deploy token.
- A CI/CD job token.

`<header-name>` is required and should be set to the proper value depending on the token type used.
See [Maven repository documentation](https://docs.khulnasoft.com/ee/user/packages/maven_repository/#edit-the-client-configuration) for more details.

#### Github

```yaml
source:
  url: https://maven.pkg.github.com/<owner>/<repository>
  credentials:
    username: $USERNAME
    token: $TOKEN
```

For more information on which `$TOKEN` can be used, see the [GitHub documentation](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-apache-maven-registry#authenticating-to-github-packages).

#### Artifactory

```yaml
source:
  url: https://<namespace>.jfrog.io/artifactory/<repository_name>
  credentials:
    username: $USERNAME
    token: $TOKEN
```

`$USERNAME` should be set to the user's email address.

To get the value for `$TOKEN`, use the [set me up instructions](https://jfrog.com/knowledge-base/the-set-me-up-option-explained/) from your Artifactory Maven repository.

### PyPI

The default image used for importing jobs is [`python:alpine`](https://hub.docker.com/_/python).

To pull packages, [`pip download`](https://pip.pypa.io/en/stable/cli/pip_download) is used.

To publish packages, [`twine upload`](https://twine.readthedocs.io/en/stable/index.html) is used. 
Because `twine` isn't present in the default image, this command is run to install it: `python -m pip install twine`.

For the credentials configuration, you need a `token` and a `username`. 
They will be used to set up a Basic Auth authentication. See the examples below.

#### PyPI Limitations

Packages are pulled and published without their dependencies. 

If the dependencies need to be imported too, they should be explicitly described in the [list of packages](#describing-packages).

#### KhulnaSoft

As a `source` package registry:

```yaml
source:
  url: https://khulnasoft.example.com/api/v4/projects/<project_id>/packages/pypi/simple
  credentials:
    username: $USERNAME
    token: $TOKEN
```

As a `destination` package registry:

```yaml
destination:
  url: https://khulnasoft.example.com/api/v4/projects/<project_id>/packages/pypi
  credentials:
    username: $USERNAME
    token: $TOKEN
```

`$TOKEN` can be one of the following [tokens](https://docs.khulnasoft.com/ee/user/packages/package_registry/supported_functionality.html#authentication-tokens),
saved as [an environment variable](#use-environment-variables-for-your-token-values):

- A personal access token.
- A deploy token.
- A CI/CD job token.

`$USERNAME` should be set to the user KhulnaSoft username.

#### Artifactory

As a `source` package registry:

```yaml
source:
  url: https://<namespace>.jfrog.io/artifactory/api/pypi/<repository_name>/simple
  credentials:
    username: $USERNAME
    token: $TOKEN
```

As a `destination` package registry:

```yaml
source:
  url: https://<namespace>.jfrog.io/artifactory/api/pypi/<repository_name>
  credentials:
    username: $USERNAME
    token: $TOKEN
```

`$USERNAME` should be set to the user's email address.

To get the value for `$TOKEN`, use the [set me up instructions](https://jfrog.com/knowledge-base/the-set-me-up-option-explained/) from your Artifactory PyPI repository.

## KhulnaSoft CI/CD

This tool generates a [KhulnaSoft child pipeline configuration](https://docs.khulnasoft.com/ee/ci/pipelines/downstream_pipelines.html#parent-child-pipelines) to execute the imports. Each import is done in a single CI/CD job. All packages of the same import type are grouped in the same stage for better readability.

### CI/CD Limitations

There are some limitations when you use KhulnaSoft child pipelines:

- The generated pipeline configuration file can't be [larger than 5 MB](https://docs.khulnasoft.com/ee/ci/pipelines/downstream_pipelines.html#dynamic-child-pipelines).
  This limit is automatically be checked by Packages Bus.
- Your subscription tier can [limit](https://docs.khulnasoft.com/ee/user/khulnasoft_com/index.html#khulnasoft-cicd)
  the number of jobs a single pipeline can host.

Due the file size limit, the amount of packages referenced in the `config.yml` file is limited too. From our testing, that limit is around `32 500` packages.

## Recommendations

### YAML Anchors

You can use [YAML anchors](https://yaml.org/spec/1.2.2/#3222-anchors-and-aliases) to make sure that your YAML file does not contain duplicates.

For example, if you have a destination registry is used multiple times, you can use a YAML
anchor to define it only once:

```yaml
import1:
  destination: &destination
    url: http://destination.registry.example/npm
    credentials:
      token: $DESTINATION_TOKEN
  -- other fields here

import2:
  destination: *destination
  source:
    url: http://source.registry.example/npm
  -- other fields here

import3:
  destination: *destination
  source:
    url: http://another_source.registry.example/npm
  -- other fields here
```

### Use environment variables for your token values

Do not put your actual tokens in the `config.yml` file. Use an environment variable for your token values.

You can set your local environment variables:

```shell
SOURCE_TOKEN=12345 pkgs_importer generate
```

## How to create a new release

1. Create a new tag. The tag name should start with `v`.
1. The pipeline for that tag will automatically build the CLI API and created a dedicated release.
