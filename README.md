#####第一步：

根据编译命令生成适于自己平台的动态库

```Bash
g++ -o DrawFont.o -c DrawFont.cpp -fPIC -I/usr/local/freetype/include/freetype2 -L/usr/local/freetype/lib -lfreetype
g++ -o DrawBrush.o -c DrawBrush.c -fPIC -I/usr/local/freetype/include/freetype2 -L/usr/local/freetype/lib -lfreetype
ar r libDrawBrush.so DrawBrush.o DrawFont.o
```

#####第二步：

修改 draw.go 里面关于 freetype 的链接路径

```
go build -o libPDraw.so -buildmode=c-shared draw.go
//会生成 libPDraw.so 和 libPDraw.h 文件
```
>  如果报`You can also manually git clone the repository to $GOPATH/src/golang.org/x/image` 请自行使用`export GOPROXY=https://mirrors.aliyun.com/goproxy/`并开启`export GO111MODULE=on`

> 如果go版本小于1.11 请自行下载`https://github.com/golang/image`并放入GOPATH中

#####第三步：

1、创建 PHP 扩展 
```bash
php-src-path/ext_skel --extname=helloworld
```
> 将golang生成的所有文件copy到 php-src-path/ext/helloworld目录下

2、在.c【`helloworld.c`】 文件中添加

```Bash
#include "libPDraw.h"
```

3、添加方法

```C
PHP_FUNCTION(MILIDrawStringToImg)
{
	char     *text;
	char     *fontfile;
	char     *imgfile;
	char     *outfile;
	char     *fontcol;
	char     *outcol;
	size_t text_len;
	size_t fontfile_len;
	size_t imgfile_len;
	size_t outfile_len;
	size_t fontcol_len;
	size_t outcol_len;
	double fontsize;
	double outsize;
	zend_long pt_w;
	zend_long pt_h;

// #ifndef FAST_ZPP  //判断是否PHP7，php7有新的语法，我这用的是老的语法
    /* Get function parameters and do error-checking. */
    if (zend_parse_parameters(ZEND_NUM_ARGS(), "ssssssddll", &text, &text_len,&fontfile, &fontfile_len,&imgfile, &imgfile_len,&outfile, &outfile_len,&fontcol, &fontcol_len,&outcol, &outcol_len,&fontsize,&outsize,&pt_w,&pt_h) == FAILURE) {
        return;
    }
// #else
//     ZEND_PARSE_PARAMETERS_START(10, 10)
//         Z_PARAM_STR(type)
//         Z_PARAM_ZVAL_EX(value, 0, 1)
//     ZEND_PARSE_PARAMETERS_END();
// #endif
    DrawStringToImg(text,fontfile,imgfile,outfile,fontsize,outsize,fontcol,outcol,pt_w,pt_h);
    return;
}
```

4、zend_function_entry 中添加

```C
PHP_FE(MILIDrawStringToImg,NULL)
```

#####第四步：

config.m4 取消 PHP_ARG_WITH 注释【`3行`】

```Bash
phpize
./configure #自己决定是否要添加 --with-php-config=php-config-path
```

修改 Makefile,将需要的 lib 和 include 加入，以及 libPDraw.so 路径和`-lPDraw`，让 PHP 扩展将 golang 的 so 文件打包进去

```
INCLUDES = -I/usr/local/php/include/php -I/usr/local/php/include/php/main -I/usr/local/php/include/php/TSRM -I/usr/local/php/include/php/Zend -I/usr/local/php/include/php/ext -I/usr/local/php/include/php/ext/date/lib -I/usr/local/freetype/include/free
LFLAGS =
LDFLAGS = -L./ -L/usr/local/freetype/lib -lPDraw -lfreetype -Wl,-rpath,/data/golang/draw/DrawFont -L/data/golang/draw/DrawFont
```
> 在INCLUDES中添加了`-I/usr/local/freetype/include/free` freetype的头文件路径

> 在LDFLAGS中 添加了 freetype的lib路径【也即是libfreetype.so的路径】以及 `-lPDraw -lfreetype`。`/data/golang/draw/DrawFont -L/data/golang/draw/DrawFont` 是libPDraw.so的路径
#####最后

make & make install
修改 php.ini 写入 extends.so

#####重启测试。。

```PHP
//测试
<?php
MILIDrawStringToImg("文字","字体文件","背景图","输出图片","字体颜色RGBA int+逗号","描边颜色RGBA int+逗号","字体大小","描边大小","文字位置x","文字位置y")
//MILIDrawStringToImg("测试22","wryhBold.ttf","background.png","outx.png","220,20,60,0","27,91,97,0",100,20,200,200);
```

![效果图](https://github.com/xyanyue/DrawFontOutline/blob/master/out.png?row=true)



参考
https://github.com/NiuStar/DrawFont/

