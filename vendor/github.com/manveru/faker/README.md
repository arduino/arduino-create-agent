# Faker for Go

![build status](https://travis-ci.org/manveru/faker.svg?branch=master)

## Usage

    package main

    import (
      "github.com/manveru/faker"
    )

    func main() {
      fake, err := faker.New("en")
      if err != nil {
	    panic(err)
	  }
      println(fake.Name())  //> "Adriana Crona"
      println(fake.Email()) //> charity.brown@fritschbotsford.biz
    }

Inspired by the ruby faker gem, which is a port of the Perl Data::Faker library.
