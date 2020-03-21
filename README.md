第一步：

根据编译命令生成适于自己平台的动态库

```Bash
g++ -o DrawFont.o -c DrawFont.cpp -fPIC -I/usr/local/freetype/include/freetype2 -L/usr/local/freetype/lib -lfreetype
g++ -o DrawBrush.o -c DrawBrush.c -fPIC -I/usr/local/freetype/include/freetype2 -L/usr/local/freetype/lib -lfreetype
ar r libDrawBrush.so DrawBrush.o DrawFont.o
```

第二步：

修改 draw.go 里面关于 freetype 的链接路径

```
go build -o libPDraw.so -buildmode=c-shared draw.go
```

第三步：

1、创建 PHP 扩展

2、在.c 文件中添加

```Bash
#include "PDraw.h"
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

// #ifndef FAST_ZPP
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

第四步：

config.m4 取消 PHP_ARG_WITH 注释

```Bash
phpize
./configure
```

修改 Makefile,将需要的 lib 和 include 加入，以及 libPDraw.so 路径和`-lPDraw`，让 PHP 扩展将 golang 的 so 文件打包进去

```
INCLUDES = -I/usr/local/php/include/php -I/usr/local/freetype/include/free -I/usr/local/php/include/php/main -I/usr/local/php/include/php/TSRM -I/usr/local/php/include/php/Zend -I/usr/local/php/include/php/ext -I/usr/local/php/include/php/ext/date/lib
LFLAGS =
LDFLAGS = -L./ -L/usr/local/freetype/lib -lPDraw -lfreetype -Wl,-rpath,/data/golang/draw/DrawFont -L/data/golang/draw/DrawFont
```

最后

make & make install
修改 php.ini 写入 extends.so
