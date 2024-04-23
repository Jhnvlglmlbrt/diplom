# diploma monitorin-ssl certificates


### to rebuild when we do tailwindcss configuration

```
npx tailwindcss-cli@latest build ./assets/app.css -o ../static/app.css
```

to run tailwindcss script named css print - npm run css


// TODO
Проверить что при повторной отправке подтверджения - текст в кнопке не исчезает
И также что в случае Email rate exceeded рендерит шаблон ошибки
// update 
походу hx-get="/confirmation" не даёт зарендерить ошибку на страницу 
x3 мб забить (так даже лучше, пользователь не увидит ошибку)

// сделать отправку уведомления на почту через gmail
// https://www.loginradius.com/blog/engineering/sending-emails-with-golang/
// https://www.youtube.com/results?search_query=golang+how+to+send+email+