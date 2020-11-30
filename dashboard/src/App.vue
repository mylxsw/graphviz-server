<template>
  <div id="app">
    <b-container fluid>
        <b-navbar type="dark" toggleable="md" variant="primary" class="mb-3" sticky>
            <b-navbar-brand href="/">Graphviz-Server <a href="https://github.com/mylxsw/graphviz-server" class="text-white" style="font-size: 30%">{{ version }}</a></b-navbar-brand>
            <b-collapse is-nav id="nav_dropdown_collapse">
                <ul class="navbar-nav flex-row ml-md-auto d-none d-md-flex"></ul>
                <b-navbar-nav>
                    <b-nav-item href="/" exact>Submit</b-nav-item>
                    <b-nav-item to="/settings">Setting</b-nav-item>
                </b-navbar-nav>
            </b-collapse>
        </b-navbar>
        <div class="main-view">
            <router-view/>
        </div>
    </b-container>
    
  </div>
</template>

<script>
    import axios from 'axios';

    export default {
        data() {
            return {
                version: 'v-0',
            }
        },
        mounted() {
            axios.get('/api/').then(response => {
                this.version = response.data.version;
            });
        },
        beforeMount() {
            axios.defaults.baseURL = this.$store.getters.serverUrl;
            let token = this.$store.getters.token;
            if (token !== "") {
                axios.defaults.headers.common['Authorization'] = "Bearer " + token;
            }
        }
    }
</script>

<style>
    .container-fluid {
        padding: 0;
    }

    .main-view {
        padding: 15px;
    }

    .th-column-width-limit {
        max-width: 300px;
    }

    @media screen and (max-width: 1366px) {
        .th-autohide-md {
            display: none;
        }
    }
    @media screen and (max-width: 768px) {
        .th-autohide-sm {
            display: none;
        }
        .search-box {
            display: none;
        }
    }

</style>
