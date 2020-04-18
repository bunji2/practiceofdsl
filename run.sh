filename=`basename $1 .txt`.go

./conv $1 $filename

go run $filename lib.go lib2.go lib3.go

#rm $filename