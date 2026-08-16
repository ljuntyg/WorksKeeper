# WorksKeeper

`<Variable>` is a variable name, `<Variables>` its plural.
`{Option 1/Option 2}` is a mandatory choice.
`[Option 1/Option 2]` is an optional choice. 
`...` is a repeatable preceding element.

## Access Control

An *object* may belong to a *subject*. A *subject* has a subset of *access rights* over an *object*.

A *subject* can only exercise the *access rights* it has over an *object*.

## Routes

A *route* reads as a grammatically correct English phrase. A *route* maps to a *page*. 

## Templates

A *page template* `<page>-page.html` maps to a *page*. A *page template* accepts *page data*.

An *entity template* `<entity>.html` maps to an *entity*. An *entity template* accepts a *template entity*.

*Page data* `<Page>Data implements Executable` may hold a *template entity*.

A *template entity* `Template<Entity> implements Templatable` holds at least the data necessary for the associated *entity template*.

## Entities

An *entity* `<Entity>` maps to a database table. The fields on an *entity* struct map to the columns in the associated database table. 

## Handlers

A *page handler* `<Page>Handler` maps to a *page*. A *page handler* holds a *page service*.

A *page handler* has at least a *handler function* called `HandleRequest()`. A *handler function* executes a *page template* with *page data* via its *page service*.

`HandleRequest()` may call the *handler functions* `handleGet()` and `handlePost()`. `handlePost()` may call other *handler functions* named `handle<Action>()`.

## Services

A *page service* `<Page>Service` is called by a *page handler*. A *page service* holds a *repository collection*.

A *page service* has at least a *service function* `GetTemplateData()` which returns *page data*.

## Repositories

A *repository* `<Entity>Repository` is called by *page services*. A *repository* may have *repository functions*.

A *repository function* runs database queries, those ending in `Tx` are transactional. A *repository function* has the following possible syntaxes.
```
Get{One<Entity>/Optional<Entity>/<Entities>}
   [By<Field>[And<Field>...]]
   [OrderBy<Field>{Ascending/Descending}]
   [LimitN][Tx]()
```
```
Insert<Entity>[Tx]()
```
```
Update<Entity>Set<Field>[And<Field>...]
      By<Field>[And<Field>...][Tx]()
```
```
Delete<Entity>By<Field>[And<Field>...][Tx]()
```

A *repository function* that runs more than one statement is instead named as a verb phrase, e.g. `AppendContentToGroupTx()`, `IncreaseContentPositionTx()`.

A *repository collection* `RepositoryCollection` holds all *repositories*.




