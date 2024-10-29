# Chrysalis

A re-implementation of Chrysalis.

Chrysalis is a tool that can be used by freelance workers to provide status
updates on their tasks.

## Entities

## User

* Has a username and password.

## Request Form

* Has a reference to a user (the creator)

## Request Form Version

* Has a reference to the underlying request form.


## Request Form Field

* Has a reference to its parent form version.

* Has an index to be sorted by.

* Tuples of (parent form ID, index) are unique.

* Has a binary toggle marking whether or not it is required.

## Checkbox Field

* Has a unique reference to a form field.

* Has an array of options.

## Radio Field

* Has a unique reference to a form field.

* Has an array of options.

## Text Field

* Has a unique reference to a form field.

* Binary toggle on whether it is single-line or multi-line.

## Number Field

* Has a unique reference to a form field.

* Has a minimum and maximum value.

* Has an optional step size.

## Filled Form

* Has a reference to a specific version of a given request form.

* Has a reference to the client's user id.

## Filled Form Field

* Has a reference to its parent filled form.

* Has an index to be sorted by.

* Tuples of (parent form ID, index) are unique.

* Binary toggle on whether or not it is filled.

## Filled Checkbox Field

* Has a reference to a filled form field.

* Has an array of selected options.

## Filled Radio Field

* Has a reference to a filled form field.

* Has one selected option.

## Filled Text Field

* Has a reference to a filled form field.

* Has a text value.

## Filled Number Field

* Has a reference to a filled form field.

* Has a numeric value.
