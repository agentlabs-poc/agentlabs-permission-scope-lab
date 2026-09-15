# Demonstration — the role record

Captured from a real run of `verify-role.sh`, then rendered into this file **from
that capture**. The SVG beside it is generated from the same capture, so the
image and this file cannot drift.

`101` lines captured, 10 commands. Reproduce with:

```sh
$(sess path)/verify-role.sh
```

## APPLICATION-MANAGED — the platform administrator ships a role

```console
abv role publish --application --name Viewer --revision 1 --permissions hrms:payroll:payslip::read
internal projection: role
id  fyb0tpp1et4w
name  Viewer
managed  application
revision  1
permissions  hrms:payroll:payslip::read
rc=0
```

## TENANT-MANAGED — a tenant administrator composes one

```console
abv role publish --name acme-payroll-approver --revision 1 --permissions hrms:payroll:payslip::read,hrms:payroll:payslip::write
internal projection: role
id  fyb0tpr4bri8
name  acme-payroll-approver
managed  tenant
revision  1
permissions  hrms:payroll:payslip::read,hrms:payroll:payslip::write
rc=0
```

## a new revision of each — --id names the role being extended

```console
abv role publish --application --id fyb0tpp1et4w --name Viewer --revision 2 --permissions hrms:payroll:payslip::read,hrms:payroll:payslip::write
internal projection: role
id  fyb0tpp1et4w
name  Viewer
managed  application
revision  2
permissions  hrms:payroll:payslip::read,hrms:payroll:payslip::write
rc=0
```

```console
abv role publish --id fyb0tpr4bri8 --name acme-payroll-approver --revision 2 --permissions hrms:payroll:payslip::read
internal projection: role
id  fyb0tpr4bri8
name  acme-payroll-approver
managed  tenant
revision  2
permissions  hrms:payroll:payslip::read
rc=0
```

## a tenant may NOT revise a role the application ships

```console
abv role publish --id fyb0tpp1et4w --name Viewer --revision 3 --permissions hrms:payroll:payslip::read
operation rejected or record not found
rc=3
```

## an id nobody issued is refused, either way

```console
abv role publish --id zzzzzzzz --name invented --revision 1 --permissions hrms:payroll:payslip::read
operation rejected or record not found
rc=3
```

## one listing, both kinds

```console
abv role list
internal projection: roles
count  5
total  5
generation  4
fi9jvxobqsxs  rev=1  tenant       payslip-reader  hrms:payroll:payslip::read
fyb0tpp1et4w  rev=1  application  Viewer  hrms:payroll:payslip::read
fyb0tpp1et4w  rev=2  application  Viewer  hrms:payroll:payslip::read,hrms:payroll:payslip::write
fyb0tpr4bri8  rev=1  tenant       acme-payroll-approver  hrms:payroll:payslip::read,hrms:payroll:payslip::write
fyb0tpr4bri8  rev=2  tenant       acme-payroll-approver  hrms:payroll:payslip::read
rc=0
```

## narrowed

```console
abv role list --managed application
internal projection: roles
count  2
total  2
generation  4
fyb0tpp1et4w  rev=1  application  Viewer  hrms:payroll:payslip::read
fyb0tpp1et4w  rev=2  application  Viewer  hrms:payroll:payslip::read,hrms:payroll:payslip::write
rc=0
```

```console
abv role list --managed tenant
internal projection: roles
count  3
total  3
generation  4
fi9jvxobqsxs  rev=1  tenant       payslip-reader  hrms:payroll:payslip::read
fyb0tpr4bri8  rev=1  tenant       acme-payroll-approver  hrms:payroll:payslip::read,hrms:payroll:payslip::write
fyb0tpr4bri8  rev=2  tenant       acme-payroll-approver  hrms:payroll:payslip::read
rc=0
```

## latest revision of each

```console
abv role list --latest
internal projection: roles
count  3
total  3
generation  4
fi9jvxobqsxs  rev=1  tenant       payslip-reader  hrms:payroll:payslip::read
fyb0tpp1et4w  rev=2  application  Viewer  hrms:payroll:payslip::read,hrms:payroll:payslip::write
fyb0tpr4bri8  rev=2  tenant       acme-payroll-approver  hrms:payroll:payslip::read
rc=0
```

## the rows

> boundary    tenant_id  key3      key4         key5             key6
> -----------  ---------  ----  ------------  ----------  ---------------------
> application             hrms  fyb0tpp1et4w  0000000001  Viewer
> application             hrms  fyb0tpp1et4w  0000000002  Viewer
> tenant       acme       hrms  fi9jvxobqsxs  0000000001  payslip-reader
> tenant       acme       hrms  fyb0tpr4bri8  0000000001  acme-payroll-approver
> tenant       acme       hrms  fyb0tpr4bri8  0000000002  acme-payroll-approver
