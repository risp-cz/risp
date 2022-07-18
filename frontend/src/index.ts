
import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import { createElement } from 'react'
import { render } from 'react-dom'
import { HashRouter } from 'react-router-dom'
import { ThemeProvider } from '@fluentui/react'
import { initializeIcons } from '@uifabric/icons'

import * as i18_en from './translations/en.json'
import * as i18_cs from './translations/cs.json'
import { Screens } from './screens'

class App {
    static init() {
        initializeIcons()

        i18n
            .use(initReactI18next)
            .init({
                fallbackLng: 'en',
                // fallbackLng: 'cs',
                defaultNS: 'risp',
                debug: true,
                resources: {
                    en: i18_en,
                    cs: i18_cs,
                },
            })
    }

    static async start() {
        App.init()

        render(
            createElement(HashRouter, {},
                createElement(ThemeProvider, {},
                    createElement(Screens, {}),
                ),
            ),
            document.getElementById('root'),
        )
    }
}

App
    .start()
    .then(() => console.log('Risp client started'))
    .catch(console.error)
