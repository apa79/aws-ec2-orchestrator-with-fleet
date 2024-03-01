# coding: utf-8
import os

print("start building...")

os.environ["GOOS"] = "linux"
os.environ["GOARCH"] = "amd64"

for item in os.listdir("src"):
    file_path = os.path.join("src", item)
    if os.path.isdir(file_path):
        old_pwd = os.getcwd()

        print(f"found lambda - {item}")
        os.chdir(file_path)

        print(f"removing existing old output binary")
        try:
            os.remove(f"../../release/binary/{item}/bootstrap")
        except OSError as e:
            pass

        print(f"building new binary")
        os.system(f"go build -o ../../release/binary/{item}/bootstrap .")

        print(f"completed")
        os.chdir(old_pwd)

old_pwd = os.getcwd()
os.chdir("terraform")
os.system(f"terraform plan")
