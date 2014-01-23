#!/usr/bin/env python3

import subprocess, json
from urllib.request import urlopen

def getSetting(name):
    try:
        raw_out = subprocess.check_output(["git", "config", "--get", name])
        return str(raw_out)
    except subprocess.CalledProcessError:
        return None

def apiGet(path, data=None, apiKey=apiKey, gitlab_url=gitlab_url, **kwargs):
    kwargs["apiKey"] = apiKey
    payload = list()
    for key, value in kwargs:
        payload.append(str(key) + '=' + str(value))
    payload = '&'.join(payload)
    #try:
    urlopen(gitlab_url + path + '?' + payload)

gitlab_url = getSetting("gitlab.url") or "http://gitlab.com/api/v3/"

if gitlab_url[-1] != '/':
    gitlab_url += '/'


