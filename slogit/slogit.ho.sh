#!/bin/bash

if [[ $#  -lt 1 ]]; then
    echo "Usage: $0 exp"
    exit
fi

EXPER=$1
cd /vlbobs/ivs/sched/
YEAR=$(date +%Y)
YEARALT=YEAR+1;
wget ftp://cddis.gsfc.nasa.gov/pub/vlbi/ivsdata/aux/$YEAR/$EXPER/$EXPER.txt
wget -r ftp://cddis.gsfc.nasa.gov/pub/vlbi/ivsdata/aux/$YEAR/$EXPER/*
if [[ ! -e $EXPER.skd ]]; then
    wget ftp://cddis.gsfc.nasa.gov/pub/vlbi/ivsdata/aux/$YEARALT/$EXPER/$EXPER.skd
    wget ftp://cddis.gsfc.nasa.gov/pub/vlbi/ivsdata/aux/$YEARALT/$EXPER/$EXPER.txt
fi
if [[ ! -e $EXPER.skd ]]; then 
    echo "I can't find the file on cddis for either $YEAR or $YEARALT. Please check the experiment's name and its availablity."
    exit
fi

scp /vlbobs/ivs/sched/$EXPER.skd oper@hobart:/usr2/sched/


ssh oper@hobart << EOF
cd /usr2/sched
echo -e "ho\n11\n7 10 1 1\n12\n3\n5\n0\n" | /usr2/fs/bin/drudg ${EXPER}.skd
mv /usr2/sched/$EXPER\ho.prc /usr2/proc/
EOF

scp oper@hobart:/tmp/sched.tmp /vlbobs/ivs/sched/${EXPER}.sum

echo Attenuation will need to be set in the ifdsx procedure of the proc file.

ssh oper@hobart 'echo "source=stow" >> /usr2/sched/'$EXPER\ho.snp





