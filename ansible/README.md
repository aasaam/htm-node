# Ansible

This is ansible installation for htm-node.

You must install minimal installation of Ubuntu 20.04 then use following steps:

## Configure

Copy sample files and edit properties:

```bash
cp group_vars/all.yml.sample group_vars/all.yml
cp hosts.yml.sample hosts.yml
```

## Download application

```bash
./download.sh
```

## Certificate

If your organization domain is `example-organization.tld`

```bash
./cert.py 'example-organization.tld'
```

## Installation

```bash
ansible-playbook -i hosts.yml playbook/install-os.yml
ansible-playbook -i hosts.yml playbook/install-node.yml
```
