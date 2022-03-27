#!/usr/bin/env python3
import itertools
import json
import os
import re
import sys
import yaml

def is_valid_hostname(hostname):
  if len(hostname) > 255 or len(hostname) < 3:
    return False
  if hostname[-1] == ".":
    hostname = hostname[:-1] # strip exactly one dot from the right, if present
  allowed = re.compile("(?!-)[A-Z\d-]{1,63}(?<!-)$", re.IGNORECASE)
  return all(allowed.match(x) for x in hostname.split("."))

base_host_name = sys.argv[1:][0]

if is_valid_hostname(base_host_name) == False:
  print("invalid host name, use ./cert your-domain.tld")
  exit(0)

csr_server_sample = None

with open('./cert/csr-servers.sample.json') as json_file:
  csr_server_sample = json.load(json_file)

with open("hosts.yml", "r") as stream:
  ansible_hosts = yaml.safe_load(stream)
  hosts = ansible_hosts['all']['hosts']
  sans = [
    base_host_name,
    '*.' + base_host_name,
    'htm.' + base_host_name,
    '*.htm.' + base_host_name,
  ]
  for host in hosts:
    props = hosts[host]
    sans.append(host)
    sans.append(props['ansible_host'])
  sans = list(k for k,_ in itertools.groupby(sorted(sans)))
  csr_server = dict(csr_server_sample)
  csr_server['hosts'] = sans

  with open('./cert/csr-servers.json', 'w') as file:
    json_string = json.dumps(csr_server, default=lambda o: o.__dict__, sort_keys=True, indent=2)
    file.write(json_string)

  os.system("&&".join([
    "cd cert",
    "./cfssl gencert -ca intermediate.pem -ca-key intermediate-key.pem -config ca-config.json -profile=server csr-servers.json | cfssljson -bare servers",
    "cat servers.pem intermediate.pem root.pem > fullchain.pem",
    "cat intermediate.pem root.pem > chain.pem",
    "cat servers-key.pem > privkey.pem",
    "cat servers.pem > cert.pem",
  ]))


