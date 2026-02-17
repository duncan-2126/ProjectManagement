# Git Push Guide

## 1. Confirm you can list remotes
```bash
 git remote -v
```

## 2. If your remote URL uses https, change it to ssh
Linux/macOS:
```bash
 git remote set-url origin git@github.com:duncan-2126/ProjectManagement.git
```

## 3. Make sure you have an SSH key added to GitHub
See the [SSH key setup guide](https://medium.com/@username/setup-ssh-key-for-github).

## 4. Test your connection
```bash
 git ls-remote git@github.com:duncan-2126/ProjectManagement.git
```

## 5. Push the branch
```bash
 git push origin feature/web-gui-automation
```
