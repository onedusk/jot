#!/bin/bash

# echo "# jot" >> README.md
git init;
git add -A;
git commit -m "first commit";
git branch -M main;
git remote add origin https://github.com/onedusk/jot.git;
git push -u origin main;
